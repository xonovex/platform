import {describe, expect, it} from "vitest";
import {streamTextDeltaLength} from "../../../../src/claude/stream-events.js";

describe("streamTextDeltaLength", () => {
  it("counts only streamed response text", () => {
    expect(
      streamTextDeltaLength({
        type: "stream_event",
        event: {
          type: "content_block_delta",
          delta: {type: "text_delta", text: "no skill"},
        },
      }),
    ).toBe(8);
    expect(streamTextDeltaLength({type: "result"})).toBe(0);
  });
});
