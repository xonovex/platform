import type { ModelProvider } from "../types.js";
import { geminiClaudeProvider } from "./gemini-claude.js";
import { geminiProvider } from "./gemini.js";
import { glmProvider } from "./glm.js";
import { gpt5CodexProvider } from "./gpt5-codex.js";

/**
 * All Claude providers
 */
export const claudeProviders: ModelProvider[] = [
  geminiProvider,
  geminiClaudeProvider,
  glmProvider,
  gpt5CodexProvider,
];
