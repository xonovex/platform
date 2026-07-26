#!/usr/bin/env node
import {main} from "./routing-owners.js";

process.exitCode = main(process.argv.slice(2));
