import { describe, expect, it, vi } from "vitest";
import { connectSystemStream, parseSSEFrames, SYSTEM_STREAM_URL, EventSourceLike, StreamEvent } from "./systemStream";

class FakeEventSource implements EventSourceLike {
  listeners = new Map<string, Array<(event: StreamEvent) => void>>();
  closed = false;

  addEventListener(type: string, listener: (event: StreamEvent) => void) {
    const list = this.listeners.get(type) ?? [];
    list.push(listener);
    this.listeners.set(type, list);
  }

  close() {
    this.closed = true;
  }

  emit(type: string, event: StreamEvent) {
    for (const listener of this.listeners.get(type) ?? []) listener(event);
  }
}

describe("system stream", () => {
  it("parses multi-frame event-stream chunks with comments, retry lines and CRLF", () => {
    const chunk = "retry: 3000\r\n\r\n: keep-alive\r\n\r\nevent: status\r\ndata: {\"a\":1}\r\n\r\nevent: error\ndata: {\"code\":\"x\"}\n\n";
    const frames = parseSSEFrames(chunk);
    expect(frames).toEqual([
      { event: "status", data: "{\"a\":1}" },
      { event: "error", data: "{\"code\":\"x\"}" },
    ]);
  });

  it("joins multi-line data fields", () => {
    const frames = parseSSEFrames("event: status\ndata: {\"a\":\ndata: 1}\n\n");
    expect(frames).toEqual([{ event: "status", data: "{\"a\":\n1}" }]);
  });

  it("delivers parsed status payloads to onStatus", () => {
    const source = new FakeEventSource();
    const onStatus = vi.fn();
    connectSystemStream({ onStatus }, { factory: () => source });
    source.emit("status", { data: JSON.stringify({ metrics: { sample: {} }, traffic_semantics: "server_counter_delta" }) });
    expect(onStatus).toHaveBeenCalledTimes(1);
    expect((onStatus.mock.calls[0][0] as { traffic_semantics: string }).traffic_semantics).toBe("server_counter_delta");
  });

  it("skips malformed status frames without throwing", () => {
    const source = new FakeEventSource();
    const onStatus = vi.fn();
    connectSystemStream({ onStatus }, { factory: () => source });
    source.emit("status", { data: "{not json" });
    expect(onStatus).not.toHaveBeenCalled();
  });

  it("maps server error frames to onError payloads", () => {
    const source = new FakeEventSource();
    const onError = vi.fn();
    connectSystemStream({ onStatus: vi.fn(), onError }, { factory: () => source });
    source.emit("error", { data: JSON.stringify({ code: "system_metrics_unavailable" }) });
    expect(onError).toHaveBeenCalledWith({ code: "system_metrics_unavailable" });
  });

  it("synthesizes a disconnect payload for transport errors without data", () => {
    const source = new FakeEventSource();
    const onError = vi.fn();
    connectSystemStream({ onStatus: vi.fn(), onError }, { factory: () => source });
    source.emit("error", {});
    expect(onError).toHaveBeenCalledWith({ code: "stream_disconnected", message: "Live stream disconnected." });
  });

  it("targets the canonical stream endpoint and closes on demand", () => {
    let requested = "";
    const source = new FakeEventSource();
    const close = connectSystemStream(
      { onStatus: vi.fn() },
      { factory: (url) => { requested = url; return source; } },
    );
    expect(requested).toBe(SYSTEM_STREAM_URL);
    close();
    expect(source.closed).toBe(true);
  });
});
