import { describe, expect, it } from "vitest";
import { encodeQR } from "./qr";

describe("local QR encoder", () => {
  it("matches a known Version 6-L mask-0 QR matrix", () => {
    const matrix = encodeQR("https://example.com/test");
    expect(matrix).toHaveLength(41);
    expect(matrix.every((row) => row.length === 41)).toBe(true);
    const rows = matrix.map((row) => row.map((cell) => (cell ? "1" : "0")).join(""));
    expect(rows[0]).toBe("11111110011100001000100010001011001111111");
    expect(rows[8]).toBe("11101111111000011001100110011111111000100");
    expect(rows[20]).toBe("01110010101010101010101010101011110111010");
    expect(rows[40]).toBe("11111110101110011001100110010101000101111");
    expect(matrix.flat().filter(Boolean)).toHaveLength(836);
  });

  it("fails closed when the subscription URL exceeds Version 6-L capacity", () => {
    expect(() => encodeQR("x".repeat(135))).toThrow(/too long/i);
  });
});
