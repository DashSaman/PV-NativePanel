package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

const maxPasswordBytes = 1024

func main() {
	reader := bufio.NewReaderSize(os.Stdin, maxPasswordBytes+2)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, os.ErrClosed) && len(line) == 0 {
		fmt.Fprintln(os.Stderr, "password helper: failed to read input")
		os.Exit(1)
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	if len(line) < 14 || len(line) > maxPasswordBytes {
		fmt.Fprintln(os.Stderr, "password helper: input length is outside policy")
		os.Exit(2)
	}
	hash, err := auth.HashPassword(line)
	line = ""
	if err != nil {
		fmt.Fprintln(os.Stderr, "password helper: hashing failed")
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, hash)
}
