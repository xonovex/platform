import cliConfig from "@xonovex/eslint-config-cli";
import portsAndAdaptersConfig from "@xonovex/eslint-config-cli/ports-and-adapters";
import {defineConfig} from "eslint/config";

export default defineConfig(cliConfig, portsAndAdaptersConfig);
