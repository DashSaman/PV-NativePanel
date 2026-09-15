import { describe, expect, it } from "vitest";
import {
  buildSmoothPath,
  fmtNum,
  donutArc,
  donutSegments,
  gaugeColor,
  gaugeDash,
  niceTicks,
  polar,
  splitFiniteRuns,
} from "./charts";

describe("monitoring chart math", () => {
  it("builds a smooth catmull-rom path through points", () => {
    const path = buildSmoothPath([
      { x: 0, y: 10 },
      { x: 50, y: 20 },
      { x: 100, y: 5 },
    ]);
    expect(path.startsWith("M0.00,10.00")).toBe(true);
    expect(path).toContain("C");
    expect(path.endsWith("100.00,5.00")).toBe(true);
  });

  it("handles degenerate point lists for smooth paths", () => {
    expect(buildSmoothPath([])).toBe("");
    expect(buildSmoothPath([{ x: 3, y: 4 }])).toBe("M3.00,4.00");
  });

  it("produces nice 1/2/5 ticks covering the data max", () => {
    expect(niceTicks(100)).toEqual([0, 50, 100]);
    expect(niceTicks(4)).toEqual([0, 1, 2, 3, 4]);
    const ticks = niceTicks(0.9);
    expect(ticks[0]).toBe(0);
    expect(ticks[ticks.length - 1]).toBeGreaterThanOrEqual(0.9);
    expect(niceTicks(0)).toEqual([0, 1]);
    expect(niceTicks(-3)).toEqual([0, 1]);
  });

  it("maps percent to the 270-degree gauge dash and threshold colors", () => {
    expect(gaugeDash(0)).toBe(0);
    expect(gaugeDash(100)).toBe(75);
    expect(gaugeDash(40)).toBeCloseTo(30);
    expect(gaugeDash(-5)).toBe(0);
    expect(gaugeDash(140)).toBe(75);
    expect(gaugeColor(10)).toBe("var(--green)");
    expect(gaugeColor(80)).toBe("var(--orange)");
    expect(gaugeColor(95)).toBe("var(--red)");
    expect(gaugeColor(Number.NaN)).toBe("var(--muted)");
  });

  it("builds donut segments with gaps and skips zero values", () => {
    const segments = donutSegments([10, 0, 10], 0);
    expect(segments).toHaveLength(2);
    expect(segments[0].a0).toBe(0);
    expect(segments[1].a1).toBeCloseTo(360);
    expect(donutSegments([0, 0])).toEqual([]);
    const gapped = donutSegments([50, 50], 4);
    expect(gapped[0].a0).toBeCloseTo(2);
    expect(gapped[0].a1 - gapped[0].a0).toBeCloseTo(176);
    expect(gapped[1].a0).toBeCloseTo(182);
    expect(gapped[1].a1).toBeCloseTo(358);
  });

  it("keeps unavailable samples as gaps instead of connecting them through zero", () => {
    expect(splitFiniteRuns([10, null, 20, 30, null, 40])).toEqual([
      [{ index: 0, value: 10 }],
      [{ index: 2, value: 20 }, { index: 3, value: 30 }],
      [{ index: 5, value: 40 }],
    ]);
  });

  it("computes donut arcs and polar coordinates consistently", () => {
    const top = polar(60, 60, 40, 0);
    expect(top.x).toBeCloseTo(60);
    expect(top.y).toBeCloseTo(20);
    const arc = donutArc(60, 60, 40, 0, 180);
    expect(arc).toContain("A40,40");
    expect(arc).toContain("0 1");
    const bigArc = donutArc(60, 60, 40, 0, 200);
    expect(bigArc).toContain("1 1");
  });
});

describe("Persian numeral formatting", () => {
  it("renders all chart/dashboard numerals with Persian digits", () => {
    expect(fmtNum(0)).toBe("۰");
    expect(fmtNum(1234)).toBe("۱٬۲۳۴");
    expect(fmtNum(70.5, 1)).toBe("۷۰٫۵");
    expect(fmtNum(Number.NaN)).toBe("—");
  });
});
