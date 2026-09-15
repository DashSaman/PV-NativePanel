import { ReactNode } from "react";

/* Inline SVG icon set (phosphor-inspired, stroke-based, zero dependency).
   All icons share a 24x24 viewBox, currentColor stroke and 1.8px width so they
   align optically with Persian text in both themes. */

const PATHS: Record<string, ReactNode> = {
  dashboard: <><rect x="3" y="3" width="7.5" height="7.5" rx="2" /><rect x="13.5" y="3" width="7.5" height="4.5" rx="2" /><rect x="13.5" y="10.5" width="7.5" height="10.5" rx="2" /><rect x="3" y="13.5" width="7.5" height="7.5" rx="2" /></>,
  users: <><circle cx="9" cy="8" r="3.4" /><path d="M3.5 19.5c0-3 2.5-5 5.5-5s5.5 2 5.5 5" /><path d="M15.5 5.4a3.2 3.2 0 0 1 0 5.9" /><path d="M17.5 14.9c1.9.7 3 2.3 3 4.6" /></>,
  plans: <><path d="M12 3 3 7.5l9 4.5 9-4.5L12 3Z" /><path d="m3 12.5 9 4.5 9-4.5" /><path d="m3 17.5 9 4.5 9-4.5" /></>,
  system: <><rect x="5" y="5" width="14" height="14" rx="3" /><rect x="9.5" y="9.5" width="5" height="5" rx="1.2" /><path d="M9 2.5v2M15 2.5v2M9 19.5v2M15 19.5v2M2.5 9h2M2.5 15h2M19.5 9h2M19.5 15h2" /></>,
  shield: <><path d="M12 3 5 5.8v5.4c0 4.3 2.9 7.6 7 9.8 4.1-2.2 7-5.5 7-9.8V5.8L12 3Z" /><path d="m9.2 11.8 2 2 3.6-4" /></>,
  logout: <><path d="M14 4h4a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-4" /><path d="m10 8-4 4 4 4" /><path d="M6 12h9" /></>,
  refresh: <><path d="M20 5v5h-5" /><path d="M4 19v-5h5" /><path d="M19.5 10a8 8 0 0 0-14.2-3M4.5 14a8 8 0 0 0 14.2 3" /></>,
  sun: <><circle cx="12" cy="12" r="4" /><path d="M12 2.5v2M12 19.5v2M2.5 12h2M19.5 12h2M5 5l1.4 1.4M17.6 17.6 19 19M19 5l-1.4 1.4M6.4 17.6 5 19" /></>,
  moon: <><path d="M20 14.5A8.5 8.5 0 0 1 9.5 4 8.5 8.5 0 1 0 20 14.5Z" /></>,
  monitor: <><rect x="3" y="4.5" width="18" height="12" rx="2" /><path d="M9 20.5h6M12 16.5v4" /></>,
  plus: <><path d="M12 5v14M5 12h14" /></>,
  search: <><circle cx="11" cy="11" r="6.5" /><path d="m16 16 4.5 4.5" /></>,
  qr: <><rect x="4" y="4" width="6.5" height="6.5" rx="1.2" /><rect x="13.5" y="4" width="6.5" height="6.5" rx="1.2" /><rect x="4" y="13.5" width="6.5" height="6.5" rx="1.2" /><path d="M13.5 13.5h2.5v2.5h-2.5zM17.5 17.5H20V20h-2.5zM13.5 20H16M20 13.5V16" /></>,
  link: <><path d="M10 14a4.2 4.2 0 0 0 6 0l3-3a4.24 4.24 0 0 0-6-6l-1.5 1.5" /><path d="M14 10a4.2 4.2 0 0 0-6 0l-3 3a4.24 4.24 0 0 0 6 6L12.5 17.5" /></>,
  copy: <><rect x="9" y="9" width="11" height="11" rx="2" /><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" transform="translate(1 0)" /></>,
  check: <><path d="m4.5 12.5 5 5 10-11" /></>,
  close: <><path d="m6 6 12 12M18 6 6 18" /></>,
  alert: <><path d="M12 3.5 2.5 20h19L12 3.5Z" /><path d="M12 10v4.5M12 17.2v.3" /></>,
  chevron: <><path d="m6 9 6 6 6-6" /></>,
  external: <><path d="M14 4h6v6" /><path d="M20 4 11 13" /><path d="M19 14v5a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h5" /></>,
  lock: <><rect x="5" y="10.5" width="14" height="10" rx="2.5" /><path d="M8 10.5V8a4 4 0 0 1 8 0v2.5" /><path d="M12 14.5v2.5" /></>,
  activity: <><path d="M3 12h4l2.5-6.5 4 13L16 12h5" /></>,
  clock: <><circle cx="12" cy="12" r="8.5" /><path d="M12 7.5V12l3 2" /></>,
  gauge: <><path d="M4 14a8 8 0 1 1 16 0" /><path d="m12 14 4-4.5" /><path d="M5.5 18.5h13" /></>,
  key: <><circle cx="8" cy="14.5" r="4.5" /><path d="m11.5 11 8-8M17 4.5 19.5 7M14.5 7.5 17 10" /></>,
  mail: <><rect x="3" y="5" width="18" height="14" rx="2.5" /><path d="m4 7 8 6 8-6" /></>,
  eye: <><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Z" /><circle cx="12" cy="12" r="3" /></>,
  "eye-off": <><path d="M4 4l16 16" /><path d="M10.6 5.9A9.8 9.8 0 0 1 12 5.8c6 0 9.5 6.2 9.5 6.2a17.6 17.6 0 0 1-2.8 3.5M6.6 6.9A16.9 16.9 0 0 0 2.5 12S6 18.2 12 18.2a9.4 9.4 0 0 0 3.4-.6" /><path d="M9.9 10.1a3 3 0 0 0 4.2 4.2" /></>,
};

export type IconName = keyof typeof PATHS;

export function Icon({ name, size = 18, className }: { name: string; size?: number; className?: string }) {
  const path = PATHS[name] ?? PATHS.dashboard;
  return (
    <svg className={className ? `icon ${className}` : "icon"} width={size} height={size} viewBox="0 0 24 24"
      fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"
      aria-hidden="true" focusable="false">
      {path}
    </svg>
  );
}

/* Lightweight SVG area sparkline (no chart dependency). Renders a smooth-ish
   polyline with a translucent gradient fill; accepts already-normalized data. */
export function Sparkline({ values, height = 34, className, title }: { values: number[]; height?: number; className?: string; title?: string }) {
  const pts = values.filter((v) => Number.isFinite(v));
  if (pts.length < 2) return <div className={`spark-empty ${className ?? ""}`} style={{ height }} aria-hidden="true" />;
  const width = 100;
  const max = Math.max(...pts, 1);
  const min = Math.min(...pts, 0);
  const span = Math.max(max - min, 0.0001);
  const step = width / (pts.length - 1);
  const coords = pts.map((v, i) => `${(i * step).toFixed(2)},${(height - 3 - ((v - min) / span) * (height - 8)).toFixed(2)}`);
  const line = `M${coords.join(" L")}`;
  const area = `${line} L${width},${height} L0,${height} Z`;
  const gradId = `spark-${Math.abs(hash(title ?? pts.join(","))).toString(36)}`;
  return (
    <svg className={className ? `spark ${className}` : "spark"} width="100%" height={height} viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none" role="img" aria-label={title ?? "روند"}>
      <defs>
        <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.28" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={area} fill={`url(#${gradId})`} stroke="none" />
      <path d={line} fill="none" stroke="currentColor" strokeWidth="1.6" vectorEffect="non-scaling-stroke" />
    </svg>
  );
}

function hash(input: string): number {
  let h = 0;
  for (let i = 0; i < input.length; i++) { h = (h * 31 + input.charCodeAt(i)) | 0; }
  return h;
}
