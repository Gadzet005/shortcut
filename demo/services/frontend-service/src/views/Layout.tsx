import type { ReactNode } from "react";

const styles = `
  *, *::before, *::after { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    background: #f5f5f7;
    color: #1a1a1a;
    min-height: 100vh;
  }
  a { color: inherit; text-decoration: none; }
  button { font-family: inherit; cursor: pointer; }

  .header {
    background: linear-gradient(90deg, #0b1f4a 0%, #1e3a8a 100%);
    color: #fff;
    padding: 14px 24px;
    position: sticky;
    top: 0;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(11, 31, 74, 0.18);
  }
  .header-inner {
    max-width: 1280px;
    margin: 0 auto;
    display: flex;
    align-items: center;
    gap: 24px;
  }
  .logo {
    font-weight: 800;
    font-size: 22px;
    letter-spacing: 0.5px;
  }
  .logo { color: #fff; }
  .header-spacer { flex: 1; }
  .header-link {
    font-size: 15px;
    font-weight: 500;
    padding: 8px 14px;
    border-radius: 8px;
    transition: background 0.15s;
  }
  .header-link:hover { background: rgba(255, 255, 255, 0.15); }

  .container {
    max-width: 1280px;
    margin: 0 auto;
    padding: 24px;
  }
  .page-title {
    font-size: 28px;
    font-weight: 700;
    margin: 8px 0 24px;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .row-card {
    background: #fff;
    border-radius: 16px;
    border: 1px solid #ececef;
    padding: 16px;
    display: grid;
    grid-template-columns: 200px 1fr 220px;
    gap: 24px;
    align-items: center;
    transition: transform 0.12s, box-shadow 0.15s;
    color: inherit;
  }
  .row-card:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.07);
    border-color: #1e3a8a33;
  }
  .row-image {
    width: 200px;
    height: 200px;
    object-fit: cover;
    border-radius: 12px;
    background: #ececef;
    display: block;
  }
  .row-info { display: flex; flex-direction: column; gap: 8px; min-width: 0; }
  .row-name {
    font-size: 18px;
    font-weight: 700;
    line-height: 1.3;
    color: #1a1a1a;
  }
  .row-description {
    font-size: 14px;
    color: #555;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .row-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 12px;
  }
  .row-price {
    font-size: 26px;
    font-weight: 800;
    color: #1a1a1a;
  }
  @media (max-width: 720px) {
    .row-card { grid-template-columns: 120px 1fr; }
    .row-image { width: 120px; height: 120px; }
    .row-actions { grid-column: 1 / -1; flex-direction: row; justify-content: space-between; align-items: center; }
  }

  .product {
    display: grid;
    grid-template-columns: minmax(280px, 480px) 1fr;
    gap: 32px;
    background: #fff;
    border-radius: 16px;
    padding: 24px;
    border: 1px solid #ececef;
  }
  .product-image {
    width: 100%;
    aspect-ratio: 1 / 1;
    object-fit: cover;
    border-radius: 12px;
    background: #ececef;
  }
  .product-name { font-size: 26px; font-weight: 700; margin: 0 0 12px; }
  .product-description { color: #555; line-height: 1.5; margin: 0 0 24px; }
  .product-price { font-size: 32px; font-weight: 800; margin: 8px 0 24px; }

  .btn-primary {
    display: inline-block;
    background: #1e3a8a;
    color: #fff;
    border: none;
    padding: 14px 28px;
    border-radius: 12px;
    font-size: 15px;
    font-weight: 600;
    transition: background 0.15s, transform 0.05s;
  }
  .btn-primary:hover { background: #1c3577; }
  .btn-primary:active { transform: scale(0.98); }
  .btn-secondary {
    background: #e6ecf8;
    color: #0b1f4a;
    border: none;
    padding: 8px 16px;
    border-radius: 10px;
    font-size: 13px;
    font-weight: 600;
    transition: background 0.15s;
  }
  .btn-secondary:hover { background: #d3dcef; }

  .section-title {
    font-size: 22px;
    font-weight: 700;
    margin: 32px 0 16px;
  }

  .cart-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 12px; }
  .cart-item {
    display: grid;
    grid-template-columns: 80px 1fr auto auto;
    gap: 16px;
    align-items: center;
    background: #fff;
    border: 1px solid #ececef;
    border-radius: 12px;
    padding: 12px 16px;
  }
  .cart-thumb {
    width: 80px;
    height: 80px;
    object-fit: cover;
    border-radius: 8px;
    background: #ececef;
  }
  .cart-name { font-weight: 600; }
  .cart-price { font-weight: 800; font-size: 18px; min-width: 100px; text-align: right; }

  .empty-state {
    background: #fff;
    border: 1px dashed #d8d8db;
    border-radius: 14px;
    padding: 48px 24px;
    text-align: center;
    color: #888;
  }

  ul.plain { list-style: none; padding: 0; margin: 0; }

  .carousel {
    display: flex;
    gap: 12px;
    overflow-x: auto;
    padding: 4px 0 16px;
    scroll-snap-type: x mandatory;
    -webkit-overflow-scrolling: touch;
  }
  .carousel::-webkit-scrollbar { height: 8px; }
  .carousel::-webkit-scrollbar-thumb { background: #d8d8db; border-radius: 4px; }
  .carousel::-webkit-scrollbar-track { background: transparent; }
  .carousel-card {
    flex: 0 0 180px;
    background: #fff;
    border: 1px solid #ececef;
    border-radius: 12px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    color: inherit;
    scroll-snap-align: start;
    transition: transform 0.12s, box-shadow 0.15s, border-color 0.15s;
  }
  .carousel-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.07);
    border-color: #1e3a8a33;
  }
  .carousel-image {
    width: 100%;
    aspect-ratio: 1 / 1;
    object-fit: cover;
    border-radius: 8px;
    background: #ececef;
  }
  .carousel-name {
    font-size: 13px;
    line-height: 1.35;
    font-weight: 500;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 36px;
  }
  .carousel-price {
    font-size: 15px;
    font-weight: 800;
  }

  .profile {
    background: #fff;
    border: 1px solid #ececef;
    border-radius: 16px;
    padding: 28px;
    display: grid;
    grid-template-columns: 240px 1fr;
    gap: 32px;
    align-items: start;
  }
  .profile-avatar {
    width: 240px;
    height: 240px;
    border-radius: 24px;
    object-fit: cover;
    background: #ececef;
    border: 4px solid #fff;
    box-shadow: 0 6px 20px rgba(11, 31, 74, 0.20);
  }
  .profile-body { min-width: 0; }
  .profile-name { font-size: 28px; font-weight: 800; margin-bottom: 8px; }
  .profile-bio { color: #555; font-size: 16px; line-height: 1.5; margin: 0 0 24px; }
  .profile-fields {
    display: grid;
    grid-template-columns: 140px 1fr;
    gap: 12px 16px;
    margin: 0;
  }
  .profile-fields dt {
    color: #888;
    font-size: 13px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .profile-fields dd {
    margin: 0;
    font-size: 15px;
    color: #1a1a1a;
    word-break: break-word;
  }
  .profile-fields dd a { color: #1e3a8a; }
  @media (max-width: 720px) {
    .profile { grid-template-columns: 1fr; }
    .profile-avatar { width: 160px; height: 160px; }
    .profile-fields { grid-template-columns: 1fr; }
    .profile-fields dt { margin-top: 8px; }
  }
`;

export function Layout({ title, children }: { title: string; children: ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{title}</title>
        <style dangerouslySetInnerHTML={{ __html: styles }} />
      </head>
      <body>
        <header className="header">
          <div className="header-inner">
            <a href="/products" className="logo">
              SHOP
            </a>
            <div className="header-spacer" />
            <a href="/products" className="header-link">
              Каталог
            </a>
            <a href="/cart" className="header-link">
              Корзина
            </a>
            <a href="/user" className="header-link">
              Профиль
            </a>
          </div>
        </header>
        <main className="container">{children}</main>
      </body>
    </html>
  );
}
