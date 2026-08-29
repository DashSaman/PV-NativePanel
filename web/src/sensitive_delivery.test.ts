import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const source = readFileSync(new URL("./ProductCustomers.tsx", import.meta.url), "utf8");

describe("one-time secret delivery", () => {
  it("does not close the replacement secret dialog after customer creation", () => {
    expect(source).not.toContain(
      "await onDone({ username: result.user.username, password: result.generated_password, subscriptionPath: result.subscription_path });\n      onClose();",
    );
  });

  it("does not close the replacement secret dialog after password rotation", () => {
    expect(source).not.toContain(
      "await onDone(result.generated_password || (generate ? undefined : password)); onClose();",
    );
  });
});
