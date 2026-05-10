package handlers

import "github.com/Gadzet005/shortcut/pkg/shortcut"

const addCartItemSQL = `
	INSERT INTO demo_cart_items (user_id, product_id, qty) VALUES ($1, $2, 1)
	ON CONFLICT (user_id, product_id) DO UPDATE
	  SET qty = demo_cart_items.qty + 1`

func AddCartItem(ctx *shortcut.Context) error {
	return cartMutation(ctx, "added", "insert cart item", addCartItemSQL)
}
