package accountingboundary

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type countingReader struct {
	reader io.Reader
	read   int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.read += int64(n)
	return n, err
}

type shortWriter struct {
	maxPerWrite int
	written     int64
}

func (w *shortWriter) Write(p []byte) (int, error) {
	accepted := len(p)
	if w.maxPerWrite >= 0 && accepted > w.maxPerWrite {
		accepted = w.maxPerWrite
	}
	w.written += int64(accepted)
	if accepted != len(p) {
		return accepted, io.ErrShortWrite
	}
	return accepted, nil
}

type successfulWriteCounter struct {
	writer  io.Writer
	written int64
}

func (w *successfulWriteCounter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.written += int64(n)
	return n, err
}

func TestSeparateHandlerBoundaryCannotInferSuccessfulUploadWritesFromBodyReads(t *testing.T) {
	payload := bytes.Repeat([]byte{0x5a}, 64)
	outer := &countingReader{reader: bytes.NewReader(payload)}
	remote := &shortWriter{maxPerWrite: 17}
	buffer := make([]byte, len(payload))

	readN, readErr := outer.Read(buffer)
	if readErr != nil {
		t.Fatalf("outer Read() error = %v", readErr)
	}
	if readN != len(payload) || outer.read != int64(len(payload)) {
		t.Fatalf("outer body observation = readN:%d counter:%d, want %d", readN, outer.read, len(payload))
	}

	writtenN, writeErr := remote.Write(buffer[:readN])
	if !errors.Is(writeErr, io.ErrShortWrite) {
		t.Fatalf("remote Write() error = %v, want io.ErrShortWrite", writeErr)
	}
	if writtenN != 17 || remote.written != 17 {
		t.Fatalf("successful remote bytes = writeN:%d counter:%d, want 17", writtenN, remote.written)
	}
	if outer.read == remote.written {
		t.Fatal("fixture failed to demonstrate the accounting boundary")
	}
	if outer.read <= remote.written {
		t.Fatalf("outer body reads = %d, successful remote writes = %d; expected attempted/read bytes to exceed successful upload bytes", outer.read, remote.written)
	}
}

func TestSeparateHandlerBoundaryCanCountItsDownloadWritesButStillCannotRecoverUploadSuccess(t *testing.T) {
	payload := bytes.Repeat([]byte{0x33}, 48)

	clientSink := &shortWriter{maxPerWrite: 23}
	downloadCounter := &successfulWriteCounter{writer: clientSink}
	downloadN, downloadErr := downloadCounter.Write(payload)
	if !errors.Is(downloadErr, io.ErrShortWrite) {
		t.Fatalf("client Write() error = %v, want io.ErrShortWrite", downloadErr)
	}
	if downloadN != 23 || downloadCounter.written != 23 || clientSink.written != 23 {
		t.Fatalf("observed successful download bytes = n:%d wrapper:%d sink:%d, want 23", downloadN, downloadCounter.written, clientSink.written)
	}

	outerUpload := &countingReader{reader: bytes.NewReader(payload)}
	remoteSink := &shortWriter{maxPerWrite: 11}
	buffer := make([]byte, len(payload))
	readN, readErr := outerUpload.Read(buffer)
	if readErr != nil {
		t.Fatalf("outer upload Read() error = %v", readErr)
	}
	remoteN, remoteErr := remoteSink.Write(buffer[:readN])
	if !errors.Is(remoteErr, io.ErrShortWrite) {
		t.Fatalf("remote upload Write() error = %v, want io.ErrShortWrite", remoteErr)
	}
	if remoteN != 11 || remoteSink.written != 11 {
		t.Fatalf("successful upload bytes = n:%d sink:%d, want 11", remoteN, remoteSink.written)
	}
	if outerUpload.read != int64(len(payload)) {
		t.Fatalf("outer upload body counter = %d, want %d", outerUpload.read, len(payload))
	}
	if outerUpload.read == remoteSink.written {
		t.Fatal("outer handler unexpectedly inferred successful client-to-remote bytes")
	}
}
