#!/usr/bin/env node
import {main} from "./validate-links.js";

process.exitCode = main(process.argv.slice(2));
