import { randomUUID } from "node:crypto";

export interface MultipartResponse {
  contentType: string;
  body: Buffer;
}

export function writeMultipartData(
  items: Record<string, string | Buffer>,
): MultipartResponse {
  const boundary = `----shortcut${randomUUID().replace(/-/g, "")}`;
  const chunks: Buffer[] = [];

  for (const [name, value] of Object.entries(items)) {
    chunks.push(
      Buffer.from(
        `--${boundary}\r\nContent-Disposition: form-data; name="${name}"\r\n\r\n`,
      ),
    );
    chunks.push(typeof value === "string" ? Buffer.from(value) : value);
    chunks.push(Buffer.from("\r\n"));
  }
  chunks.push(Buffer.from(`--${boundary}--\r\n`));

  return {
    contentType: `multipart/form-data; boundary=${boundary}`,
    body: Buffer.concat(chunks),
  };
}

export interface HttpResponseItem {
  status_code: number;
  headers: Record<string, string[]>;
  body: string;
}

export function buildHttpResponseItem(
  statusCode: number,
  contentType: string,
  body: string,
): HttpResponseItem {
  return {
    status_code: statusCode,
    headers: { "Content-Type": [contentType] },
    body: Buffer.from(body, "utf-8").toString("base64"),
  };
}
