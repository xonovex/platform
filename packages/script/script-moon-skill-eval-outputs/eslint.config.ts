import baseConfig from "@xonovex/eslint-config-cli";
import {defineConfig} from "eslint/config";
import {workspaceRules} from "../../../eslint.workspace.ts";

export default defineConfig(baseConfig, ...workspaceRules(import.meta.dirname));
