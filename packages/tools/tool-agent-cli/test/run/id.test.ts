import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  AGENT_ID_ENV,
  generateAgentId,
  getAgentIdEnv,
  PARENT_AGENT_ID_ENV,
} from "../../src/run/id.js";

describe("id", () => {
  describe("generateAgentId", () => {
    it("should generate a 7 character hex string", () => {
      const id = generateAgentId();

      expect(id).toHaveLength(7);
      expect(id).toMatch(/^[0-9a-f]{7}$/);
    });

    it("should generate unique IDs", () => {
      const ids = new Set<string>();
      for (let i = 0; i < 100; i++) {
        ids.add(generateAgentId());
      }

      // All 100 IDs should be unique
      expect(ids.size).toBe(100);
    });

    it("should only contain lowercase hex characters", () => {
      for (let i = 0; i < 50; i++) {
        const id = generateAgentId();
        expect(id).toMatch(/^[0-9a-f]+$/);
        expect(id).not.toMatch(/[A-F]/);
      }
    });
  });

  describe("getAgentIdEnv", () => {
    const originalEnv = process.env;

    beforeEach(() => {
      process.env = { ...originalEnv };
    });

    afterEach(() => {
      process.env = originalEnv;
    });

    it("should set XONOVEX_AGENT_ID", () => {
      const env = getAgentIdEnv("abc1234");

      expect(env[AGENT_ID_ENV]).toBe("abc1234");
    });

    it("should not set XONOVEX_PARENT_AGENT_ID when no existing agent ID", () => {
      process.env = { ...process.env };
      process.env.XONOVEX_AGENT_ID = undefined;

      const env = getAgentIdEnv("abc1234");

      expect(env[PARENT_AGENT_ID_ENV]).toBeUndefined();
    });

    it("should set XONOVEX_PARENT_AGENT_ID when existing agent ID in environment", () => {
      process.env[AGENT_ID_ENV] = "parent123";

      const env = getAgentIdEnv("child456");

      expect(env[AGENT_ID_ENV]).toBe("child456");
      expect(env[PARENT_AGENT_ID_ENV]).toBe("parent123");
    });

    it("should support nested agent hierarchies", () => {
      // Simulate grandparent -> parent -> child
      process.env[AGENT_ID_ENV] = "grandparent";

      const parentEnv = getAgentIdEnv("parent");
      expect(parentEnv[AGENT_ID_ENV]).toBe("parent");
      expect(parentEnv[PARENT_AGENT_ID_ENV]).toBe("grandparent");

      // Now simulate parent spawning child
      process.env[AGENT_ID_ENV] = "parent";

      const childEnv = getAgentIdEnv("child");
      expect(childEnv[AGENT_ID_ENV]).toBe("child");
      expect(childEnv[PARENT_AGENT_ID_ENV]).toBe("parent");
    });
  });

  describe("environment variable names", () => {
    it("should export correct env var names", () => {
      expect(AGENT_ID_ENV).toBe("XONOVEX_AGENT_ID");
      expect(PARENT_AGENT_ID_ENV).toBe("XONOVEX_PARENT_AGENT_ID");
    });
  });
});
