import {describe, expect, it} from "vitest";
import {main} from "./index.js";

describe("main", () => {
  it("fails cleanly when the query file is missing", async () => {
    await expect(main(["missing-queries.json", "testing"])).resolves.toBe(2);
  });
});
