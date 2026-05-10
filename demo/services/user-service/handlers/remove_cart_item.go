package handlers

import "github.com/Gadzet005/shortcut/pkg/shortcut"

const removeCartItemSQL = `
	UPDATE demo_cart_items SET qty = GREATEST(qty - 1, 0)
	WHERE user_id = $1 AND product_id = $2`

func RemoveCartItem(ctx *shortcut.Context) error {
	return cartMutation(ctx, "removed", "remove cart item", removeCartItemSQL)
}
