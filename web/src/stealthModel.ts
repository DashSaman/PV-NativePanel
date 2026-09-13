// Pure reveal state machine for the stealth login (R8 / UI-001).
//
// Stealth contract: at rest no box, label, or placeholder is visible. A field
// materializes when the pointer enters its zone (hover) OR it receives
// keyboard focus. Focus is sticky against hover-leave so Tab navigation and
// password managers keep working. Failed-login messages render inside a
// reserved slot so the layout never shifts (no visual tell).

export type StealthField = "email" | "password" | "totp";

export type StealthState = {
  hovered: StealthField | null;
  focused: StealthField | null;
};

export type StealthEvent =
  | { type: "hover-enter"; field: StealthField }
  | { type: "hover-leave"; field: StealthField }
  | { type: "focus"; field: StealthField }
  | { type: "blur"; field: StealthField };

export const initialStealthState: StealthState = { hovered: null, focused: null };

export function reduceStealth(state: StealthState, event: StealthEvent): StealthState {
  switch (event.type) {
    case "hover-enter":
      return { ...state, hovered: event.field };
    case "hover-leave":
      return state.hovered === event.field ? { ...state, hovered: null } : state;
    case "focus":
      return { ...state, focused: event.field };
    case "blur":
      return state.focused === event.field ? { ...state, focused: null } : state;
    default: {
      const never: never = event;
      return never;
    }
  }
}

export function isRevealed(state: StealthState, field: StealthField): boolean {
  return state.hovered === field || state.focused === field;
}

// Reserved height (px) of the message slot under the form. The slot is always
// rendered, so a failed-login message never changes layout.
export const MESSAGE_SLOT_MIN_PX = 30;
