import type { ClientConfig, Query, Response, RunOptions } from "./types.js";

export class ShortcutClient {
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;

  constructor(cfg: ClientConfig) {
    if (!cfg.baseUrl) {
      throw new Error("shortcutclient: empty baseUrl");
    }
    this.baseUrl = cfg.baseUrl.replace(/\/+$/, "");
    this.fetchImpl = cfg.fetch ?? globalThis.fetch.bind(globalThis);
  }

  async run(
    graph: string,
    body: string | Uint8Array | null = null,
    opts: RunOptions = {},
  ): Promise<Response> {
    if (!graph) {
      throw new Error("shortcutclient: empty graph name");
    }

    const url = buildUrl(this.baseUrl, graph, opts.query);
    const headers = new Headers(opts.headers);
    if (!headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    const signal = withTimeout(opts.signal, opts.timeout);

    const resp = await this.fetchImpl(url, {
      method: opts.method ?? "POST",
      headers,
      body: body as BodyInit | null,
      signal,
    });

    const buf = new Uint8Array(await resp.arrayBuffer());

    return {
      status: resp.status,
      headers: resp.headers,
      body: buf,
      json<T = unknown>(): T {
        return JSON.parse(new TextDecoder().decode(buf)) as T;
      },
      text(): string {
        return new TextDecoder().decode(buf);
      },
    };
  }
}

export function createClient(cfg: ClientConfig): ShortcutClient {
  return new ShortcutClient(cfg);
}

function buildUrl(baseUrl: string, graph: string, query?: Query): string {
  const path = graph.replace(/^\/+/, "");
  const url = new URL(`${baseUrl}/run/${path}`);
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (Array.isArray(value)) {
        for (const v of value) {
          url.searchParams.append(key, String(v));
        }
      } else {
        url.searchParams.append(key, String(value));
      }
    }
  }
  return url.toString();
}

function withTimeout(
  signal: AbortSignal | undefined,
  timeout: number | undefined,
): AbortSignal | undefined {
  if (!timeout || timeout <= 0) {
    return signal;
  }
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeout);
  if (signal) {
    if (signal.aborted) {
      controller.abort();
    } else {
      signal.addEventListener("abort", () => controller.abort(), { once: true });
    }
  }
  controller.signal.addEventListener("abort", () => clearTimeout(timer), { once: true });
  return controller.signal;
}
