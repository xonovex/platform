#!/usr/bin/env node
// npm links a workspace bin only when its target exists at install time, and
// dist/ is generated, so the bin points at this committed file instead.
import "./dist/src/index.js";
