import {describe, expect, it} from "vitest";
import {
  DEFAULT_TMUX_CONFIG,
  type TmuxConfig,
} from "../../src/run/wrapper/tmux/types.js";

describe("tmux", () => {
  describe("DEFAULT_TMUX_CONFIG", () => {
    it("should have correct default values", () => {
      expect(DEFAULT_TMUX_CONFIG.enabled).toBe(false);
      expect(DEFAULT_TMUX_CONFIG.detach).toBe(false);
      expect(DEFAULT_TMUX_CONFIG.attachExisting).toBe(true);
      expect(DEFAULT_TMUX_CONFIG.sessionName).toBeUndefined();
      expect(DEFAULT_TMUX_CONFIG.windowName).toBeUndefined();
    });
  });

  describe("TmuxConfig type", () => {
    it("should accept valid config", () => {
      const config: TmuxConfig = {
        enabled: true,
        sessionName: "my-session",
        windowName: "agent",
        detach: false,
        attachExisting: true,
      };

      expect(config.enabled).toBe(true);
      expect(config.sessionName).toBe("my-session");
      expect(config.windowName).toBe("agent");
      expect(config.detach).toBe(false);
      expect(config.attachExisting).toBe(true);
    });

    it("should allow optional fields", () => {
      const config: TmuxConfig = {
        enabled: true,
        detach: false,
        attachExisting: false,
      };

      expect(config.sessionName).toBeUndefined();
      expect(config.windowName).toBeUndefined();
    });
  });
});
