import { optionalJSONItem, requireJSONItem } from "@shortcut/client";
import { makeRender } from "../lib/render.js";
import { ProductList } from "../views/ProductList.js";
import { ProductPage } from "../views/ProductPage.js";
import { CartPage } from "../views/CartPage.js";
import { UserPage, type User } from "../views/UserPage.js";
import type { Product } from "../views/types.js";

export const renderProductList = makeRender({
  parse: (b) => ({ products: requireJSONItem<Product[]>(b, "products") }),
  view: ({ products }) => <ProductList products={products} />,
});

export const renderProductPage = makeRender({
  parse: (b) => ({
    product: requireJSONItem<Product>(b, "product"),
    recommendations: optionalJSONItem<Product[]>(b, "recommendations", []),
  }),
  view: ({ product, recommendations }) => (
    <ProductPage product={product} recommendations={recommendations} />
  ),
});

export const renderCart = makeRender({
  parse: (b) => ({ products: requireJSONItem<Product[]>(b, "products") }),
  view: ({ products }) => <CartPage products={products} />,
});

export const renderUser = makeRender({
  parse: (b) => ({ user: requireJSONItem<User>(b, "user") }),
  view: ({ user }) => <UserPage user={user} />,
});
