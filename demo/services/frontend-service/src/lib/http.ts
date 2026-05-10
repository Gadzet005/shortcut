import type { Request, Response } from "express";
import type { Query } from "@shortcut/client";

const SKIPPED_HEADERS = new Set([
  "content-length",
  "transfer-encoding",
  "connection",
  "set-cookie",
]);

export interface UpstreamResponse {
  status: number;
  headers: Headers;
  body: Uint8Array;
}

export function parseExpressQuery(req: Request, base: Query = {}): Query {
  const out: Query = { ...base };
  for (const [key, value] of Object.entries(req.query)) {
    if (typeof value === "string") {
      out[key] = value;
    } else if (Array.isArray(value)) {
      out[key] = value.filter((v): v is string => typeof v === "string");
    }
  }
  return out;
}

export function parseFormBody(
  req: Request,
  base: Record<string, string> = {},
): Record<string, string> {
  const out: Record<string, string> = { ...base };
  for (const [key, value] of Object.entries(req.body as Record<string, unknown>)) {
    if (typeof value === "string") out[key] = value;
  }
  return out;
}

export function relay(res: Response, upstream: UpstreamResponse): void {
  res.status(upstream.status);
  upstream.headers.forEach((value, key) => {
    if (SKIPPED_HEADERS.has(key.toLowerCase())) return;
    res.setHeader(key, value);
  });
  res.send(Buffer.from(upstream.body));
}

export function sendUpstreamError(res: Response, err: unknown): void {
  const message = err instanceof Error ? err.message : String(err);
  res.status(502).type("text/plain").send(`upstream error: ${message}`);
}

export function readCookie(req: Request, name: string): string | undefined {
  const header = req.headers.cookie;
  if (!header) return undefined;
  for (const part of header.split(";")) {
    const eq = part.indexOf("=");
    if (eq < 0) continue;
    const key = part.slice(0, eq).trim();
    if (key === name) return decodeURIComponent(part.slice(eq + 1).trim());
  }
  return undefined;
}
