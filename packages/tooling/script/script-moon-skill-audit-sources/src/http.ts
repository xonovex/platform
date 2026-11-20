export interface HttpResponse {
  readonly status: number;
  readonly lastModified: string | undefined;
  /** The body, read only when readBodyWhen accepted the status. */
  readonly body: Uint8Array | undefined;
}

export interface HttpRequest {
  readonly timeoutSeconds: number;
  readonly readBodyWhen: (status: number) => boolean;
}

/** Reads a URL; a network failure or a timeout rejects. */
export interface HttpClient {
  readonly get: (url: string, request: HttpRequest) => Promise<HttpResponse>;
}
