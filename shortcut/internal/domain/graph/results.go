package graph

func newGraphResults() graphResults {
	return make(graphResults)
}

type graphResults map[NodeID]map[ItemID]Item

func (g graphResults) Add(nodeID NodeID, itemID ItemID, item Item) {
	if g[nodeID] == nil {
		g[nodeID] = make(map[ItemID]Item)
	}
	g[nodeID][itemID] = item
}

func (g graphResults) Get(nodeID NodeID, itemID ItemID) (Item, bool) {
	nodeItems, ok := g[nodeID]
	if !ok {
		return Item{}, false
	}
	item, ok := nodeItems[itemID]
	return item, ok
}

func (g graphResults) GetAny(nodeID NodeID) (Item, bool) {
	nodeItems, ok := g[nodeID]
	if !ok {
		return Item{}, false
	}
	for _, item := range nodeItems {
		return item, true
	}
	return Item{}, false
}
