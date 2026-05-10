package handlers

type CatalogEntry struct {
	ID       string
	Category string
}

var catalog []CatalogEntry

func SetCatalog(entries []CatalogEntry) {
	catalog = make([]CatalogEntry, len(entries))
	copy(catalog, entries)
}

func categoryOf(productID string) string {
	for _, e := range catalog {
		if e.ID == productID {
			return e.Category
		}
	}
	return ""
}

func idsByCategory(category string) []string {
	out := make([]string, 0, len(catalog))
	for _, e := range catalog {
		if e.Category == category {
			out = append(out, e.ID)
		}
	}
	return out
}

func allIDs() []string {
	out := make([]string, 0, len(catalog))
	for _, e := range catalog {
		out = append(out, e.ID)
	}
	return out
}
