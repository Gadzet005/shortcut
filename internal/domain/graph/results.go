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
func (g graphResults) GetAll(nodeID NodeID) map[ItemID]Item {
	nodeItems, ok := g[nodeID]
	if !ok {
		return nil
	}
	return nodeItems
}
