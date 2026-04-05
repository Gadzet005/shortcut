package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Store namespace has three routes at the same path with different HTTP methods:
//
//	GET    /store/items → list_items graph
//	POST   /store/items → create_item graph  (validate-item → save-item)
//	DELETE /store/items → clear_items graph
//
// The mock service keeps items in memory, so tests are run sequentially to
// observe state changes across requests.

func TestStore(t *testing.T) {
	// Ensure a clean slate before this test suite.
	deleted := storeDeleteItems(t)
	require.GreaterOrEqual(t, deleted.DeletedCount, 0)

	t.Run("list is empty initially", func(t *testing.T) {
		resp := storeListItems(t)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Empty(t, resp.Items)
	})

	t.Run("POST creates first item and returns it with generated id", func(t *testing.T) {
		resp := storeCreateItem(t, "Apple", 1.5)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Equal(t, 1, resp.Item.ID)
		require.Equal(t, "Apple", resp.Item.Name)
		require.InDelta(t, 1.5, resp.Item.Price, 0.001)
		require.NotEmpty(t, resp.Item.CreatedAt)
	})

	t.Run("GET reflects persisted state after POST", func(t *testing.T) {
		resp := storeListItems(t)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, resp.Items, 1)
		require.Equal(t, "Apple", resp.Items[0].Name)
	})

	t.Run("POST creates second item with incremented id", func(t *testing.T) {
		resp := storeCreateItem(t, "Banana", 2.0)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Equal(t, 2, resp.Item.ID)
		require.Equal(t, "Banana", resp.Item.Name)
	})

	t.Run("GET shows both items in insertion order", func(t *testing.T) {
		resp := storeListItems(t)
		require.Len(t, resp.Items, 2)
		require.Equal(t, "Apple", resp.Items[0].Name)
		require.Equal(t, "Banana", resp.Items[1].Name)
	})

	t.Run("DELETE clears all items and returns correct count", func(t *testing.T) {
		resp := storeDeleteItems(t)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, 2, resp.DeletedCount)
	})

	t.Run("GET is empty again after DELETE", func(t *testing.T) {
		resp := storeListItems(t)
		require.Empty(t, resp.Items)
	})

	t.Run("IDs reset after DELETE — new item gets id 1", func(t *testing.T) {
		resp := storeCreateItem(t, "Cherry", 3.0)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Equal(t, 1, resp.Item.ID)
	})

	// Clean up after the test so other test runs start fresh.
	storeDeleteItems(t)
}

func TestStoreValidation(t *testing.T) {
	t.Run("missing name returns 400", func(t *testing.T) {
		resp := storeCreateItem(t, "", 1.5)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("negative price returns 400", func(t *testing.T) {
		resp := storeCreateItem(t, "Apple", -1.0)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("zero price returns 400", func(t *testing.T) {
		resp := storeCreateItem(t, "Apple", 0)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid price string returns 400", func(t *testing.T) {
		resp := storeCreateItemRaw(t, "Apple", "not-a-number")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestStoreMethodRouting(t *testing.T) {
	t.Run("PUT to /store/items returns 404 (method not registered)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, shortcutURL+"/run/store/store/items", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("PATCH to /store/items returns 404 (method not registered)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, shortcutURL+"/run/store/store/items", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

type storeItem struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

type storeListResponse struct {
	StatusCode int
	Items      []storeItem
}

type storeCreateResponse struct {
	StatusCode int
	Item       storeItem
}

type storeDeleteResponse struct {
	StatusCode   int
	DeletedCount int
}

func storeListItems(t *testing.T) storeListResponse {
	t.Helper()
	resp, err := http.Get(shortcutURL + "/run/store/store/items")
	require.NoError(t, err)
	defer resp.Body.Close()

	result := storeListResponse{StatusCode: resp.StatusCode}
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Items))
	}
	return result
}

func storeCreateItem(t *testing.T, name string, price float64) storeCreateResponse {
	t.Helper()
	return storeCreateItemRaw(t, name, fmt.Sprintf("%g", price))
}

func storeCreateItemRaw(t *testing.T, name, price string) storeCreateResponse {
	t.Helper()
	url := shortcutURL + "/run/store/store/items"
	sep := "?"
	if name != "" {
		url += sep + "name=" + name
		sep = "&"
	}
	if price != "" {
		url += sep + "price=" + price
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := storeCreateResponse{StatusCode: resp.StatusCode}
	if resp.StatusCode == http.StatusCreated {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Item))
	}
	return result
}

func storeDeleteItems(t *testing.T) storeDeleteResponse {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, shortcutURL+"/run/store/store/items", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := storeDeleteResponse{StatusCode: resp.StatusCode}
	if resp.StatusCode == http.StatusOK {
		body := struct {
			DeletedCount int `json:"deleted_count"`
		}{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		result.DeletedCount = body.DeletedCount
	}
	return result
}
