import { Layout } from "./Layout.js";
import { formatPrice } from "./format.js";
import type { Product } from "./types.js";

export function CartPage({ products }: { products: Product[] }) {
  const total = products.reduce((sum, p) => sum + p.price, 0);

  return (
    <Layout title="SHOP - Корзина">
      <h1 className="page-title">Корзина</h1>
      {products.length === 0 ? (
        <div className="empty-state">
          Корзина пуста.{" "}
          <a href="/products" style={{ color: "#1e3a8a", fontWeight: 600 }}>
            Перейти в каталог
          </a>
        </div>
      ) : (
        <>
          <ul className="cart-list">
            {products.map((p) => (
              <li key={p.id} className="cart-item">
                <img className="cart-thumb" src={p.imageUrl} alt={p.name} />
                <a href={`/product?product_id=${p.id}`} className="cart-name">
                  {p.name}
                </a>
                <div className="cart-price">{formatPrice(p.price)}</div>
                <form method="post" action="/cart/remove">
                  <input type="hidden" name="product_id" value={p.id} />
                  <button type="submit" className="btn-secondary">
                    Удалить
                  </button>
                </form>
              </li>
            ))}
          </ul>
          <div
            style={{
              marginTop: 24,
              padding: "20px 24px",
              background: "#fff",
              border: "1px solid #ececef",
              borderRadius: 14,
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <span style={{ fontSize: 18, fontWeight: 600 }}>Итого</span>
            <span style={{ fontSize: 26, fontWeight: 800 }}>{formatPrice(total)}</span>
          </div>
        </>
      )}
    </Layout>
  );
}
