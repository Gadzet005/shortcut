import express from "express";
import { createClient } from "@shortcut/client";
import type { Config } from "./config.js";
import { Graphs, GRAPH } from "./lib/graphs.js";
import { createUserSession } from "./lib/session.js";
import { proxyForm, proxyGet } from "./lib/proxy.js";
import {
  renderCart,
  renderProductList,
  renderProductPage,
  renderUser,
} from "./handlers/renders.js";

export async function startServer(cfg: Config): Promise<void> {
  const app = express();
  app.use(express.urlencoded({ extended: false, limit: "10mb" }));

  const session = await createUserSession();
  const graphs = new Graphs(createClient({ baseUrl: cfg.shortcut["base-url"] }));
  const deps = { graphs, session };

  app.get("/health", (_req, res) => {
    res.json({ status: "ok" });
  });
  app.get("/", (_req, res) => {
    res.redirect(302, "/products");
  });

  app.get("/products", proxyGet(deps, GRAPH.productList, { limit: "10" }));
  app.get("/product",  proxyGet(deps, GRAPH.productPage, { limit: "8" }));
  app.get("/cart",     proxyGet(deps, GRAPH.cart));
  app.get("/user",     proxyGet(deps, GRAPH.userProfile));
  app.post("/cart/add",    proxyForm(deps, GRAPH.cartAdd,    "/cart"));
  app.post("/cart/remove", proxyForm(deps, GRAPH.cartRemove, "/cart"));

  app.post("/render-product-list", renderProductList);
  app.post("/render-product-page", renderProductPage);
  app.post("/render-cart",         renderCart);
  app.post("/render-user",         renderUser);

  const port = cfg.http.port;
  await new Promise<void>((resolve) => {
    app.listen(port, () => {
      console.log(JSON.stringify({
        msg: "http server started",
        port,
        userIds: session.availableIds,
      }));
      resolve();
    });
  });
}
