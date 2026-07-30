#!/usr/bin/env node
import {main} from "./routing.js";

process.exitCode = main(process.argv.slice(2));
