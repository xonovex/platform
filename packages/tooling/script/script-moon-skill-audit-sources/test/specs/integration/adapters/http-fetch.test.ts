import {createServer, type Server} from "node:http";
import type {AddressInfo} from "node:net";
import {afterAll, beforeAll, describe, expect, it} from "vitest";
import {fetchHttpClient} from "../../../../src/adapters/http-fetch.js";

describe("fetchHttpClient", () => {
  let server: Server;
  let base = "";

  beforeAll(async () => {
    server = createServer((request, response) => {
      if (request.url === "/slow") return;
      response.statusCode = request.url === "/gone" ? 404 : 200;
      response.setHeader("Last-Modified", "Tue, 01 Jan 2030 00:00:00 GMT");
      response.end(`agent ${request.headers["user-agent"] ?? ""}`);
    });
    await new Promise<void>((resolve) => {
      server.listen(0, "127.0.0.1", resolve);
    });
    base = `http://127.0.0.1:${String((server.address() as AddressInfo).port)}`;
  });

  afterAll(() => {
    server.closeAllConnections();
    server.close();
  });

  it("reads the status, the date and the body it is asked for", async () => {
    const response = await fetchHttpClient.get(`${base}/page`, {
      timeoutSeconds: 5,
      readBodyWhen: (status) => status === 200,
    });

    expect(response.status).toBe(200);
    expect(response.lastModified).toBe("Tue, 01 Jan 2030 00:00:00 GMT");
    expect(new TextDecoder().decode(response.body)).toBe(
      "agent skill-source-audit/1",
    );
  });

  it("leaves the body unread when the status does not call for it", async () => {
    const response = await fetchHttpClient.get(`${base}/gone`, {
      timeoutSeconds: 5,
      readBodyWhen: (status) => status === 200,
    });

    expect(response).toMatchObject({status: 404, body: undefined});
  });

  it("rejects a request that outlasts its timeout", async () => {
    await expect(
      fetchHttpClient.get(`${base}/slow`, {
        timeoutSeconds: 0.2,
        readBodyWhen: () => true,
      }),
    ).rejects.toThrow("aborted");
  });
});
