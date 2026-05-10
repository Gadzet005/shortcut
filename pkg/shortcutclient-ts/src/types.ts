export interface ClientConfig {
  baseUrl: string;
  fetch?: typeof fetch;
}

export type QueryValue = string | number | boolean;
export type Query = Record<string, QueryValue | QueryValue[]>;

export interface RunOptions {
  method?: string;
  headers?: HeadersInit;
  query?: Query;
  signal?: AbortSignal;
  timeout?: number;
}

export interface Response {
  status: number;
  headers: Headers;
  body: Uint8Array;
  json<T = unknown>(): T;
  text(): string;
}
