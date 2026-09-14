export type SSEFrame = { event: string; data: string };

// Minimal structural surface of EventSource we depend on. Keeping it
// structural lets tests inject a fake without DOM globals.
export interface StreamEvent {
  data?: string;
}

export interface EventSourceLike {
  addEventListener(type: string, listener: (event: StreamEvent) => void): void;
  close(): void;
}

export type EventSourceFactory = (url: string) => EventSourceLike;

export type SystemStreamHandlers = {
  onStatus: (payload: unknown) => void;
  onError?: (payload: unknown) => void;
};

export const SYSTEM_STREAM_URL = "/api/v1/system/stream";

// parseSSEFrames turns a chunk of an event-stream body into frames.
// It tolerates CRLF line endings, comment/retry lines, and multi-line data.
export function parseSSEFrames(chunk: string): SSEFrame[] {
  const frames: SSEFrame[] = [];
  let event = "";
  let data = "";
  const flush = () => {
    if (event !== "" || data !== "") {
      frames.push({ event, data });
      event = "";
      data = "";
    }
  };
  for (const rawLine of chunk.split("\n")) {
    const line = rawLine.replace(/\r$/, "");
    if (line.startsWith("event: ")) {
      event = line.slice("event: ".length);
    } else if (line.startsWith("data: ")) {
      data += (data === "" ? "" : "\n") + line.slice("data: ".length);
    } else if (line === "") {
      flush();
    }
    // comment (":...") and "retry:" lines are intentionally ignored
  }
  return frames;
}

function parsePayload(raw: string): unknown {
  return JSON.parse(raw);
}

// connectSystemStream subscribes to the live system metrics stream. Frames
// named "status" carry the same envelope as the polling endpoint, so
// normalizeSystemStatus can consume them unchanged. Server-sent "error"
// frames carry a JSON failure envelope; transport-level disconnects surface
// as a synthetic stream_disconnected payload. Returns a close() function.
export function connectSystemStream(
  handlers: SystemStreamHandlers,
  options?: { url?: string; factory?: EventSourceFactory },
): () => void {
  const factory: EventSourceFactory =
    options?.factory ??
    ((url: string) => new EventSource(url) as unknown as EventSourceLike);
  const source = factory(options?.url ?? SYSTEM_STREAM_URL);

  source.addEventListener("status", (event) => {
    if (typeof event.data !== "string") return;
    try {
      handlers.onStatus(parsePayload(event.data));
    } catch {
      // Malformed frame: skip rather than tear down the stream.
    }
  });
  source.addEventListener("error", (event) => {
    if (typeof event.data === "string" && event.data !== "") {
      try {
        handlers.onError?.(parsePayload(event.data));
        return;
      } catch {
        // fall through to the synthetic disconnect payload
      }
    }
    handlers.onError?.({ code: "stream_disconnected", message: "Live stream disconnected." });
  });

  return () => source.close();
}
