import { Layout } from "./Layout.js";
import { formatPrice } from "./format.js";
import type { Product } from "./types.js";

export function ProductList({ products }: { products: Product[] }) {
  return (
    <Layout title="SHOP - Каталог">
      <h1 className="page-title">
        Каталог{" "}
        <span style={{ color: "#888", fontWeight: 500, fontSize: 18 }}>
          {products.length} товаров
        </span>
      </h1>
      {products.length === 0 ? (
        <div className="empty-state">Пока ничего не найдено.</div>
      ) : (
        <div className="list">
          {products.map((p) => (
            <a key={p.id} className="row-card" href={`/product?product_id=${p.id}`}>
              <img className="row-image" src={p.imageUrl} alt={p.name} loading="lazy" />
              <div className="row-info">
                <div className="row-name">{p.name}</div>
                <div className="row-description">{p.description}</div>
              </div>
              <div className="row-actions">
                <div className="row-price">{formatPrice(p.price)}</div>
                <span className="btn-primary">Подробнее</span>
              </div>
            </a>
          ))}
        </div>
      )}
    </Layout>
  );
}
