import { describe, expect, it } from "vitest";
import {
  MESSAGE_SLOT_MIN_PX,
  initialStealthState,
  isRevealed,
  reduceStealth,
} from "./stealthModel";

describe("stealth login reveal machine", () => {
  it("keeps every field hidden at rest", () => {
    for (const field of ["email", "password", "totp"] as const) {
      expect(isRevealed(initialStealthState, field)).toBe(false);
    }
  });

  it("reveals only the hovered field", () => {
    const s = reduceStealth(initialStealthState, { type: "hover-enter", field: "email" });
    expect(isRevealed(s, "email")).toBe(true);
    expect(isRevealed(s, "password")).toBe(false);
    expect(isRevealed(s, "totp")).toBe(false);
  });

  it("hover-leave hides the field again", () => {
    let s = reduceStealth(initialStealthState, { type: "hover-enter", field: "password" });
    s = reduceStealth(s, { type: "hover-leave", field: "password" });
    expect(isRevealed(s, "password")).toBe(false);
  });

  it("hover-leave for another field does not disturb the revealed one", () => {
    let s = reduceStealth(initialStealthState, { type: "hover-enter", field: "email" });
    s = reduceStealth(s, { type: "hover-leave", field: "password" });
    expect(isRevealed(s, "email")).toBe(true);
  });

  it("keyboard focus reveals and is sticky against hover-leave", () => {
    let s = reduceStealth(initialStealthState, { type: "focus", field: "password" });
    expect(isRevealed(s, "password")).toBe(true);
    s = reduceStealth(s, { type: "hover-enter", field: "password" });
    s = reduceStealth(s, { type: "hover-leave", field: "password" });
    expect(isRevealed(s, "password")).toBe(true); // focus keeps it visible
  });

  it("blur hides the focused field unless the pointer is still there", () => {
    let s = reduceStealth(initialStealthState, { type: "hover-enter", field: "email" });
    s = reduceStealth(s, { type: "focus", field: "email" });
    s = reduceStealth(s, { type: "blur", field: "email" });
    expect(isRevealed(s, "email")).toBe(true); // hover holds it open
    s = reduceStealth(s, { type: "hover-leave", field: "email" });
    expect(isRevealed(s, "email")).toBe(false);
  });

  it("moving the pointer between fields hands over the hover state", () => {
    let s = reduceStealth(initialStealthState, { type: "hover-enter", field: "email" });
    s = reduceStealth(s, { type: "hover-enter", field: "password" });
    expect(isRevealed(s, "email")).toBe(false);
    expect(isRevealed(s, "password")).toBe(true);
  });

  it("reserves the message slot so failed logins never shift layout", () => {
    expect(MESSAGE_SLOT_MIN_PX).toBeGreaterThan(0);
  });
});
