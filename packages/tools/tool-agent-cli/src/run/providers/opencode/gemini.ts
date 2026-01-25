import type { ModelProvider } from "../types.js";

/**
 * Google Gemini provider configuration for OpenCode
 * Uses --model flag to specify the model
 */
export const geminiProvider: ModelProvider = {
  name: "gemini",
  displayName: "Google Gemini",
  agentType: "opencode",
  environment: {},
  cliArgs: ["--model", "google/gemini-2.5-pro"],
};
