import type { RequestHandler } from "express";
import type { ReactElement } from "react";
import { renderToString } from "react-dom/server";
import { respondWithHtml } from "@shortcut/client";

export interface RenderOptions<T> {
  parse: (body: Record<string, string>) => T;
  view: (data: T) => ReactElement;
}

export function makeRender<T>(opts: RenderOptions<T>): RequestHandler {
  return (req, res) => {
    const body = req.body as Record<string, string>;
    let data: T;
    try {
      data = opts.parse(body);
    } catch (err) {
      const message = err instanceof Error ? err.message : "invalid input";
      res.status(400).json({ error: message });
      return;
    }
    const html = "<!DOCTYPE html>" + renderToString(opts.view(data));
    respondWithHtml(res, html);
  };
}
