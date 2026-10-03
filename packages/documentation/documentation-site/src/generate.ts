import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {generateSite} from "./site.js";

const site = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const root = resolve(site, "../../..");

try {
  const result = generateSite(root, site, process.argv.includes("--check"));
  console.log(
    "Generated " +
      String(result.pages) +
      " documentation pages and " +
      String(result.catalog.entries.length) +
      " component entries; updated " +
      String(result.changedReadmes.length) +
      " installation sections.",
  );
} catch (error: unknown) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}
