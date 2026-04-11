import type { TraceResponse } from "./types";

export async function fetchTrace(requestId: string): Promise<TraceResponse> {
  const resp = await fetch(`/trace/${encodeURIComponent(requestId)}`);
  if (!resp.ok) {
    const body = await resp.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${resp.status}`);
  }
  return resp.json();
}
