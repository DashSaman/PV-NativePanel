// Design tokens for the PVNaive command-center UI (R9 "Midnight Glass Ops").
//
// Methodology: ui-ux-pro-max-skill (typography scale, 8pt spacing grid,
// semantic color roles, WCAG AA contrast, motion rules). These tokens mirror
// the CSS custom properties in styles.css/system.css; component code must not
// invent raw values outside this file.

export const spacing = {
  xxs: "4px",
  xs: "8px",
  sm: "12px",
  md: "16px",
  lg: "24px",
  xl: "32px",
  xxl: "48px",
} as const;

export const radius = {
  sm: "10px",
  md: "14px",
  lg: "20px",
  pill: "999px",
} as const;

export const typography = {
  display: { size: "28px", lineHeight: "1.35", weight: 800 },
  title: { size: "20px", lineHeight: "1.4", weight: 800 },
  body: { size: "14px", lineHeight: "1.7", weight: 400 },
  caption: { size: "12px", lineHeight: "1.6", weight: 400 },
  mono: { family: '"JetBrains Mono","Vazirmatn",ui-monospace,monospace', feature: "tabular-nums" },
} as const;

// Motion durations per the ui-ux-pro-max skill motion scale. The stealth
// login uses `reveal` for hover/focus materialization.
export const motion = {
  fast: "120ms",
  reveal: "180ms",
  slow: "240ms",
  ease: "cubic-bezier(0.2, 0.6, 0.2, 1)",
} as const;

// Semantic palette. Values mirror styles.css custom properties so the
// stealth layer and the panel shell render from one palette.
// "Private Gold on Midnight Glass" (R10): deep-space navy canvas, frosted
// glass surfaces, gold/amber brand accent (Private Network logo). Chart data
// series keep teal/violet so data colors never collide with the gold chrome.
export const palette = {
  bgDeep: "#04070f",
  surface: "rgba(17, 27, 49, 0.60)",
  surfaceRaised: "rgba(148, 197, 255, 0.06)",
  border: "rgba(128, 160, 220, 0.14)",
  borderLuminous: "rgba(255, 255, 255, 0.09)",
  text: "#eaf1ff",
  textDim: "#8ea2c7",
  accent: "#f5b62e",
  accentSoft: "rgba(245, 182, 46, 0.16)",
  accent2: "#e8990c",
  danger: "#fb7185",
  ok: "#34d399",
} as const;

export const designTokens = { spacing, radius, typography, motion, palette } as const;
