import { Layout } from "./Layout.js";
import { formatPrice } from "./format.js";
import type { Product } from "./types.js";

export function ProductPage({
  product,
  recommendations,
}: {
  product: Product;
  recommendations: Product[];
}) {
  return (
    <Layout title={`SHOP - ${product.name}`}>
      <div className="product">
        <img className="product-image" src={product.imageUrl} alt={product.name} />
        <div>
          <h1 className="product-name">{product.name}</h1>
          <div className="product-price">{formatPrice(product.price)}</div>
          <p className="product-description">{product.description}</p>
          <form method="post" action="/cart/add">
            <input type="hidden" name="product_id" value={product.id} />
            <button type="submit" className="btn-primary">
              Добавить в корзину
            </button>
          </form>
        </div>
      </div>

      <h2 className="section-title">С этим товаром покупают</h2>
      {recommendations.length === 0 ? (
        <div className="empty-state">Рекомендаций пока нет.</div>
      ) : (
        <div className="carousel">
          {recommendations.map((p) => (
            <a key={p.id} className="carousel-card" href={`/product?product_id=${p.id}`}>
              <img className="carousel-image" src={p.imageUrl} alt={p.name} loading="lazy" />
              <div className="carousel-name">{p.name}</div>
              <div className="carousel-price">{formatPrice(p.price)}</div>
            </a>
          ))}
        </div>
      )}
    </Layout>
  );
}
