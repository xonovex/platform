#!/usr/bin/env node
import {main} from "./validate-drift.js";

process.exitCode = main(process.argv.slice(2));
