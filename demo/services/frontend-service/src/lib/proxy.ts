import type { RequestHandler } from "express";
import type { Query } from "@shortcut/client";
import type { Graphs, GraphName } from "./graphs.js";
import type { UserSession } from "./session.js";
import { parseExpressQuery, parseFormBody, relay, sendUpstreamError } from "./http.js";

export interface ProxyDeps {
  graphs: Graphs;
  session: UserSession;
}

export function proxyGet(
  deps: ProxyDeps,
  graph: GraphName,
  defaults: Record<string, string> = {},
): RequestHandler {
  return async (req, res) => {
    const query = parseExpressQuery(req, {
      user_id: deps.session.pickUserId(req, res),
      ...defaults,
    });
    try {
      relay(res, await deps.graphs.fetch(graph, query));
    } catch (err) {
      sendUpstreamError(res, err);
    }
  };
}

export function proxyForm(
  deps: ProxyDeps,
  graph: GraphName,
  redirectTo: string,
): RequestHandler {
  return async (req, res) => {
    const fields = parseFormBody(req, { user_id: deps.session.pickUserId(req, res) });
    try {
      const upstream = await deps.graphs.invoke(graph, fields as Query);
      if (upstream.status >= 200 && upstream.status < 400) {
        res.redirect(303, redirectTo);
        return;
      }
      relay(res, upstream);
    } catch (err) {
      sendUpstreamError(res, err);
    }
  };
}
