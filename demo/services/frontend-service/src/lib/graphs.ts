import type { ShortcutClient, Query, Response as ShortcutResponse } from "@shortcut/client";

export const GRAPH = {
  productList: "shop/products",
  productPage: "shop/product",
  cart:        "shop/cart",
  userProfile: "shop/user",
  cartAdd:     "shop/cart/add",
  cartRemove:  "shop/cart/remove",
} as const;

export type GraphName = (typeof GRAPH)[keyof typeof GRAPH];

export class Graphs {
  constructor(private readonly client: ShortcutClient) {}

  fetch(name: GraphName, query: Query): Promise<ShortcutResponse> {
    return this.client.run(name, null, { method: "GET", query });
  }

  invoke(name: GraphName, query: Query): Promise<ShortcutResponse> {
    return this.client.run(name, null, { method: "POST", query });
  }
}
