import { buildHttpResponseItem, writeMultipartData } from "./multipart.js";

export interface ResponseLike {
  status(code: number): this;
  type(contentType: string): this;
  send(body: Buffer | Uint8Array | string): this;
}

export function respondWithHtml(
  res: ResponseLike,
  html: string,
  status = 200,
): void {
  const item = buildHttpResponseItem(status, "text/html; charset=utf-8", html);
  const { contentType, body } = writeMultipartData({
    http_response: JSON.stringify(item),
  });
  res.status(200).type(contentType).send(body);
}

export function respondWithItems(
  res: ResponseLike,
  items: Record<string, string | Buffer>,
): void {
  const { contentType, body } = writeMultipartData(items);
  res.status(200).type(contentType).send(body);
}

export function requireJSONItem<T>(body: Record<string, string>, name: string): T {
  const raw = body[name];
  if (!raw) throw new Error(`missing ${name} item`);
  try {
    return JSON.parse(raw) as T;
  } catch {
    throw new Error(`invalid ${name} json`);
  }
}

export function optionalJSONItem<T>(
  body: Record<string, string>,
  name: string,
  fallback: T,
): T {
  const raw = body[name];
  if (!raw) return fallback;
  try {
    return JSON.parse(raw) as T;
  } catch {
    throw new Error(`invalid ${name} json`);
  }
}
