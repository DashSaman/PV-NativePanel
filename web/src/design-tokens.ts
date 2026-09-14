// Design tokens for the PVNaive command-center UI (R8).
//
// Methodology: ui-ux-pro-max-skill (typography scale, 8pt spacing grid,
// semantic color roles, WCAG AA contrast, motion rules). These tokens are the
// single source of truth consumed by CSS custom properties and TS styles;
// component code must not invent raw values outside this file.

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
  display: { size: "28px", lineHeight: "1.35", weight: 700 },
  title: { size: "20px", lineHeight: "1.4", weight: 700 },
  body: { size: "14px", lineHeight: "1.7", weight: 400 },
  caption: { size: "12px", lineHeight: "1.6", weight: 400 },
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
export const palette = {
  bgDeep: "#0b0d12",
  surface: "rgba(255, 255, 255, 0.04)",
  surfaceRaised: "rgba(255, 255, 255, 0.07)",
  border: "rgba(255, 255, 255, 0.10)",
  borderLuminous: "rgba(255, 255, 255, 0.18)",
  text: "#e8eaf0",
  textDim: "#9aa3b2",
  accent: "#f5b942",
  accentSoft: "rgba(245, 185, 66, 0.16)",
  danger: "#ff6b6b",
  ok: "#4ade80",
} as const;

export const designTokens = { spacing, radius, typography, motion, palette } as const;
