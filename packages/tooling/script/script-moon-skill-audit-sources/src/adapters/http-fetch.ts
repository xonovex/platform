import type {HttpClient} from "../http.js";

/** Reads URLs with the global fetch, aborting a request that outlasts its timeout. */
export const fetchHttpClient: HttpClient = {
  get: async (url, {readBodyWhen, timeoutSeconds}) => {
    const controller = new AbortController();
    const timer = setTimeout(() => {
      controller.abort();
    }, timeoutSeconds * 1000);
    try {
      const response = await fetch(url, {
        headers: {"User-Agent": "skill-source-audit/1"},
        signal: controller.signal,
      });
      return {
        status: response.status,
        lastModified: response.headers.get("Last-Modified") ?? undefined,
        body: readBodyWhen(response.status)
          ? new Uint8Array(await response.arrayBuffer())
          : undefined,
      };
    } finally {
      clearTimeout(timer);
    }
  },
};
