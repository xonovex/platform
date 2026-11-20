import {describe, expect, it, vi} from "vitest";
import {aggregateContentSha256, fetchStatus} from "../../../src/fetch-check.js";
import type {HttpClient} from "../../../src/http.js";

const client = (get: HttpClient["get"]): HttpClient => ({get});

describe("fetchStatus", () => {
  it("names the Last-Modified date and hashes the body of a success", async () => {
    const get = vi.fn<HttpClient["get"]>().mockResolvedValue({
      status: 200,
      lastModified: "Tue, 01 Jan 2030 00:00:00 GMT",
      body: new TextEncoder().encode("body"),
    });

    await expect(
      fetchStatus(client(get), "https://example.com"),
    ).resolves.toEqual({
      status: "ok",
      detail: "HTTP 200, Last-Modified: Tue, 01 Jan 2030 00:00:00 GMT",
      content_sha256:
        "230d8358dc8e8890b4c58deeb62912ee2f20357ae92a5cc861b98e68fe31acb5",
    });
    const request = get.mock.calls[0]?.[1];
    expect(
      [200, 404, 403, 500].map((status) => request?.readBodyWhen(status)),
    ).toEqual([true, false, false, false]);
  });

  it("reports a missing page without retrying", async () => {
    const get = vi
      .fn<HttpClient["get"]>()
      .mockResolvedValue({status: 404, lastModified: "x", body: undefined});

    await expect(
      fetchStatus(client(get), "https://example.com"),
    ).resolves.toEqual({status: "missing", detail: "HTTP 404"});
    expect(get).toHaveBeenCalledTimes(1);
  });

  it("retries an unreachable URL once and reports the second failure", async () => {
    const get = vi
      .fn<HttpClient["get"]>()
      .mockRejectedValueOnce(new Error("reset"))
      .mockRejectedValueOnce("refused");

    await expect(
      fetchStatus(client(get), "https://example.com"),
    ).resolves.toEqual({status: "error", detail: "unreachable: refused"});
  });
});

describe("aggregateContentSha256", () => {
  it("returns null unless every URL returned content", () => {
    expect(
      aggregateContentSha256(
        ["a", "b"],
        [{status: "ok", detail: "", content_sha256: "x"}],
      ),
    ).toBeNull();
    expect(
      aggregateContentSha256(["a"], [{status: "missing", detail: "HTTP 404"}]),
    ).toBeNull();
  });
});
