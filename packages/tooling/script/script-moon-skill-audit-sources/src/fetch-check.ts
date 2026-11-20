import {createHash} from "node:crypto";
import type {HttpClient} from "./http.js";

export interface FetchReport {
  status: string;
  detail: string;
  content_sha256?: string;
}

const TIMEOUT_SECONDS = 15;

const isSuccess = (status: number): boolean => status > 0 && status < 400;

const fetchStatusOnce = async (
  http: HttpClient,
  url: string,
): Promise<FetchReport> => {
  try {
    const response = await http.get(url, {
      timeoutSeconds: TIMEOUT_SECONDS,
      readBodyWhen: (status) =>
        isSuccess(status) && ![401, 403, 429].includes(status),
    });
    const code = response.status;
    const detail =
      `HTTP ${String(code)}` +
      (response.lastModified
        ? `, Last-Modified: ${response.lastModified}`
        : "");
    if (code === 404) {
      return {status: "missing", detail: `HTTP ${String(code)}`};
    }
    if ([401, 403, 429].includes(code)) {
      return {status: "restricted", detail};
    }
    if (!isSuccess(code)) return {status: "error", detail};
    return {
      status: "ok",
      detail,
      content_sha256: createHash("sha256")
        .update(response.body ?? new Uint8Array())
        .digest("hex"),
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {status: "error", detail: `unreachable: ${message}`};
  }
};

/** Checks that a URL still resolves, retrying a failure once. */
export const fetchStatus = async (
  http: HttpClient,
  url: string,
): Promise<FetchReport> => {
  const report = await fetchStatusOnce(http, url);
  return report.status === "error" ? fetchStatusOnce(http, url) : report;
};

/**
 * The digest of the ordered fetched content of a source, or null when any URL
 * did not return content.
 */
export const aggregateContentSha256 = (
  urls: readonly string[],
  fetches: readonly FetchReport[],
): string | null => {
  if (
    urls.length !== fetches.length ||
    fetches.some(
      (fetch) => fetch.status !== "ok" || fetch.content_sha256 === undefined,
    )
  ) {
    return null;
  }
  const digest = createHash("sha256");
  for (const [index, url] of urls.entries()) {
    digest.update(url);
    digest.update("\0");
    digest.update(fetches[index]?.content_sha256 ?? "");
    digest.update("\n");
  }
  return digest.digest("hex");
};
