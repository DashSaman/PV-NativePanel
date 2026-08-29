#!/usr/bin/env python3
from __future__ import annotations

import argparse
import concurrent.futures
import json
import statistics
import time
import urllib.request


def one(url: str, timeout: float) -> tuple[bool, float, int]:
    started = time.perf_counter()
    try:
        with urllib.request.urlopen(url, timeout=timeout) as response:
            response.read(4096)
            ok = 200 <= response.status < 300
            status = response.status
    except Exception:
        ok, status = False, 0
    return ok, (time.perf_counter() - started) * 1000, status


def percentile(values: list[float], p: float) -> float:
    if not values:
        return 0.0
    ordered = sorted(values)
    index = min(len(ordered) - 1, max(0, round((len(ordered) - 1) * p)))
    return ordered[index]


def main() -> int:
    parser = argparse.ArgumentParser(description="Bounded PVNaive control-plane HTTP rehearsal; not a capacity ceiling benchmark.")
    parser.add_argument("--url", default="http://127.0.0.1:8080/api/v1/health/live")
    parser.add_argument("--requests", type=int, default=300)
    parser.add_argument("--concurrency", type=int, default=12)
    parser.add_argument("--timeout", type=float, default=3.0)
    args = parser.parse_args()
    if not 1 <= args.requests <= 5000 or not 1 <= args.concurrency <= 100:
        raise SystemExit("requests/concurrency outside safe rehearsal bounds")

    started = time.perf_counter()
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as pool:
        results = list(pool.map(lambda _: one(args.url, args.timeout), range(args.requests)))
    duration = time.perf_counter() - started
    latencies = [latency for ok, latency, _ in results if ok]
    failures = sum(1 for ok, _, _ in results if not ok)
    report = {
        "kind": "control_plane_rehearsal_not_capacity_ceiling",
        "url": args.url,
        "requests": args.requests,
        "concurrency": args.concurrency,
        "success": args.requests - failures,
        "failures": failures,
        "duration_seconds": round(duration, 3),
        "observed_requests_per_second": round(args.requests / duration, 2) if duration else 0,
        "latency_ms_p50": round(statistics.median(latencies), 2) if latencies else 0,
        "latency_ms_p95": round(percentile(latencies, 0.95), 2),
        "latency_ms_max": round(max(latencies), 2) if latencies else 0,
    }
    print(json.dumps(report, sort_keys=True))
    return 0 if failures == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
