import {describe, expect, it} from "vitest";
import {main} from "./index.js";

describe("main", () => {
  it("renders help without auditing the filesystem", async () => {
    await expect(main(["--help"])).resolves.toBe(0);
  });
});
