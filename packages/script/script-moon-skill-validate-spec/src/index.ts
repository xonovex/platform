#!/usr/bin/env node
import {main} from "./validate-skill.js";

process.exitCode = main(process.argv.slice(2));
