import { useMemo } from "react";

/* ============================================================
   PVNaive monitoring chart kit — pure SVG, zero dependencies.
   Patterned after professional monitoring consoles (Grafana /
   Netdata): live area series, radial gauges, segmented donuts.
   All math lives in exported pure functions so vitest can pin
   the geometry without a DOM. Rendering only consumes real
   server data — no component here invents values.
   ============================================================ */

export type Point = { x: number; y: number };

/** Catmull-Rom → cubic Bézier: smooth monitoring-style curve through points. */
export function buildSmoothPath(points: Point[]): string {
  if (points.length === 0) return "";
  if (points.length === 1) return `M${points[0].x.toFixed(2)},${points[0].y.toFixed(2)}`;
  let d = `M${points[0].x.toFixed(2)},${points[0].y.toFixed(2)}`;
  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i - 1] ?? points[i];
    const p1 = points[i];
    const p2 = points[i + 1];
    const p3 = points[i + 2] ?? p2;
    const c1x = p1.x + (p2.x - p0.x) / 6;
    const c1y = p1.y + (p2.y - p0.y) / 6;
    const c2x = p2.x - (p3.x - p1.x) / 6;
    const c2y = p2.y - (p3.y - p1.y) / 6;
    d += ` C${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p2.x.toFixed(2)},${p2.y.toFixed(2)}`;
  }
  return d;
}

/** "Nice" y-axis ticks (1/2/5 × 10ⁿ) covering [0, max]. */
export function niceTicks(max: number, count = 4): number[] {
  if (!Number.isFinite(max) || max <= 0) return [0, 1];
  const rawStep = max / Math.max(1, count);
  const magnitude = Math.pow(10, Math.floor(Math.log10(rawStep)));
  let step = magnitude * 10;
  for (const m of [1, 2, 5, 10]) {
    if (m * magnitude >= rawStep) { step = m * magnitude; break; }
  }
  const ticks: number[] = [];
  for (let v = 0; v <= max + step * 0.001; v += step) ticks.push(Number(v.toFixed(10)));
  const top = Number((Math.ceil(max / step) * step).toFixed(10));
  if (ticks[ticks.length - 1] < top) ticks.push(top);
  return ticks;
}

/** Arc flag math for 270° radial gauges (start 135°, sweep clockwise). */
export function gaugeDash(percent: number): number {
  const clamped = Math.max(0, Math.min(100, Number.isFinite(percent) ? percent : 0));
  return clamped * 0.75;
}

export function gaugeColor(percent: number): string {
  if (!Number.isFinite(percent)) return "var(--muted)";
  if (percent >= 90) return "var(--red)";
  if (percent >= 75) return "var(--orange)";
  return "var(--green)";
}

/** Donut segment angles in degrees with a small gap between segments. */
export function donutSegments(values: number[], gapDeg = 3): Array<{ a0: number; a1: number }> {
  const finite = values.map((v) => (Number.isFinite(v) && v > 0 ? v : 0));
  const total = finite.reduce((sum, v) => sum + v, 0);
  if (total <= 0) return [];
  const usable = Math.max(0, 360 - gapDeg * finite.filter((v) => v > 0).length);
  const segments: Array<{ a0: number; a1: number }> = [];
  let cursor = gapDeg / 2;
  for (const value of finite) {
    if (value <= 0) continue;
    const sweep = (value / total) * usable;
    segments.push({ a0: cursor, a1: cursor + sweep });
    cursor += sweep + gapDeg;
  }
  return segments;
}

/** Polar → cartesian helper for donut arcs. */
export function polar(cx: number, cy: number, r: number, deg: number): Point {
  const rad = ((deg - 90) * Math.PI) / 180;
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
}

export function donutArc(cx: number, cy: number, r: number, a0: number, a1: number): string {
  const start = polar(cx, cy, r, a0);
  const end = polar(cx, cy, r, a1);
  const large = a1 - a0 > 180 ? 1 : 0;
  return `M${start.x.toFixed(3)},${start.y.toFixed(3)} A${r},${r} 0 ${large} 1 ${end.x.toFixed(3)},${end.y.toFixed(3)}`;
}

// Chart numerals use Latin digits + JetBrains Mono (tabular) — the NOC
// convention. Persian digits render in a fallback font (Vazirmatn) whose
// metrics break mono axis labels; Latin digits keep tick columns aligned.
export const fmtNum = (value: number, digits = 0): string =>
  Number.isFinite(value)
    ? value.toLocaleString("en-US", { maximumFractionDigits: digits })
    : "—";

const faFormat = (value: number, digits = 0): string => fmtNum(value, digits);

function stableId(input: string): string {
  let h = 0;
  for (let i = 0; i < input.length; i++) { h = (h * 31 + input.charCodeAt(i)) | 0; }
  return Math.abs(h).toString(36);
}

export type SeriesValue = number | null;
export type Series = { name: string; color: string; values: SeriesValue[] };

export type FiniteRunPoint = { index: number; value: number };

/** Split a nullable series into contiguous finite runs so Unknown renders as a gap. */
export function splitFiniteRuns(values: SeriesValue[]): FiniteRunPoint[][] {
  const runs: FiniteRunPoint[][] = [];
  let current: FiniteRunPoint[] = [];
  for (let index = 0; index < values.length; index++) {
    const value = values[index];
    if (typeof value === "number" && Number.isFinite(value)) {
      current.push({ index, value });
      continue;
    }
    if (current.length) runs.push(current);
    current = [];
  }
  if (current.length) runs.push(current);
  return runs;
}

/** Live multi-series area chart with grid, nice ticks and smooth curves. */
export function LiveAreaChart({ series, height = 190, formatValue = (n) => faFormat(n), ariaLabel }: {
  series: Series[];
  height?: number;
  formatValue?: (value: number) => string;
  ariaLabel: string;
}) {
  const width = 300;
  const padTop = 8;
  const padBottom = 16;
  const plotH = height - padTop - padBottom;

  const { paths, ticks, max } = useMemo(() => {
    const all = series.flatMap((s) => s.values).filter((v): v is number => typeof v === "number" && Number.isFinite(v));
    const dataMax = all.length ? Math.max(...all) : 1;
    const top = niceTicks(dataMax).slice(-1)[0] ?? dataMax;
    const localTicks = niceTicks(top);
    const span = Math.max(top, 0.0001);
    const built = series.flatMap((s, seriesIndex) => {
      const denominator = Math.max(1, s.values.length - 1);
      return splitFiniteRuns(s.values).map((run, runIndex) => {
        const pts: Point[] = run.map(({ index, value }) => ({
          x: (index / denominator) * width,
          y: padTop + plotH - (Math.max(0, Math.min(value, span)) / span) * plotH,
        }));
        const line = buildSmoothPath(pts);
        const firstX = pts[0]?.x ?? 0;
        const lastX = pts[pts.length - 1]?.x ?? firstX;
        return {
          key: `${seriesIndex}-${runIndex}`,
          color: s.color,
          gradientIndex: seriesIndex,
          line,
          area: pts.length ? `${line} L${lastX},${padTop + plotH} L${firstX},${padTop + plotH} Z` : "",
        };
      });
    });
    return { paths: built, ticks: localTicks, max: span };
  }, [series, plotH]);

  const count = Math.max(1, ...series.map((s) => s.values.length - 1), 1);

  return (
    <svg className="monitor-chart" width="100%" height={height} viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none" role="img" aria-label={ariaLabel}>
      <defs>
        {series.map((s, i) => (
          <linearGradient key={s.name} id={`mgrad-${stableId(ariaLabel)}-${i}`} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={s.color} stopOpacity="0.45" />
            <stop offset="55%" stopColor={s.color} stopOpacity="0.14" />
            <stop offset="100%" stopColor={s.color} stopOpacity="0.03" />
          </linearGradient>
        ))}
      </defs>
      {ticks.map((t) => {
        const y = padTop + plotH - (t / max) * plotH;
        return (
          <g key={t}>
            <line x1="0" x2={width} y1={y} y2={y} className="monitor-grid" />
            <text x={width - 2} y={y - 2.5} className="monitor-tick" textAnchor="end">{formatValue(t)}</text>
          </g>
        );
      })}
      <line x1={width - 0.5} x2={width - 0.5} y1={padTop} y2={padTop + plotH} className="monitor-cursor" vectorEffect="non-scaling-stroke" />
      {paths.map((path) => <path key={`a-${path.key}`} d={path.area} fill={`url(#mgrad-${stableId(ariaLabel)}-${path.gradientIndex})`} stroke="none" />)}
      {paths.map((path) => (
        <path key={`l-${path.key}`} d={path.line} fill="none" stroke={path.color} strokeWidth="2.25"
          vectorEffect="non-scaling-stroke" strokeLinecap="round" className="monitor-line" />
      ))}
      <line x1="0" x2={width} y1={padTop + plotH} y2={padTop + plotH} className="monitor-axis" />
      <text x="1" y={height - 4} className="monitor-tick" textAnchor="start">−{faFormat(count)}s</text>
      <text x={width - 1} y={height - 4} className="monitor-tick" textAnchor="end">اکنون</text>
    </svg>
  );
}

/** 270° radial gauge with threshold coloring and center label. */
export function RadialGauge({ percent, label, caption, valueText }: {
  percent: number;
  label: string;
  caption?: string;
  valueText?: string;
}) {
  const color = gaugeColor(percent);
  const dash = gaugeDash(percent);
  return (
    <div className="monitor-gauge" role="img" aria-label={`${label}: ${valueText ?? faFormat(percent, 1)}%`}>
      <svg width="100%" height="112" viewBox="0 0 120 84">
        <g transform="rotate(135 60 46)">
          <circle cx="60" cy="46" r="34" pathLength="100" className="monitor-gauge-track" strokeDasharray="75 25" />
          <circle cx="60" cy="46" r="34" pathLength="100" className="monitor-gauge-value"
            stroke={color} strokeDasharray={`${dash.toFixed(2)} 100`} />
        </g>
        <text x="60" y="52" textAnchor="middle" className="monitor-gauge-num" fill={color}>
          {valueText ?? `${faFormat(percent, 1)}٪`}
        </text>
      </svg>
      <span className="monitor-gauge-label">{label}</span>
      {/* Usage caption lives OUTSIDE the arc: no collision with the gauge
          stroke, full-width mono text, readable at a glance. */}
      {caption && <span className="monitor-gauge-caption">{caption}</span>}
    </div>
  );
}

/** Segmented donut with center total (real data only). */
export function DonutChart({ values, colors, center, caption, ariaLabel, size = 168 }: {
  values: number[];
  colors: string[];
  center: string;
  caption: string;
  ariaLabel: string;
  size?: number;
}) {
  const segments = useMemo(() => donutSegments(values), [values]);
  return (
    <div className="monitor-donut" style={{ width: size, height: size }} role="img" aria-label={ariaLabel}>
      <svg width={size} height={size} viewBox="0 0 120 120">
        <circle cx="60" cy="60" r="40" className="monitor-donut-track" />
        {segments.map((seg, i) => (
          <path key={i} d={donutArc(60, 60, 40, seg.a0, seg.a1)} className="monitor-donut-seg"
            stroke={colors[i] ?? "var(--muted)"} />
        ))}
      </svg>
      <div className="monitor-donut-center"><strong>{center}</strong><span>{caption}</span></div>
    </div>
  );
}
