package graph

import "errors"

type dagNode struct {
	id   NodeID
	next []NodeID
}

type dag map[NodeID]dagNode

func convertToDAG(g Graph) (dag, error) {
	d := dag{
		InputNodeID: dagNode{
			id:   InputNodeID,
			next: []NodeID{},
		},
	}

	for _, node := range g.Nodes {
		d[node.ID()] = dagNode{
			id:   node.ID(),
			next: []NodeID{},
		}
	}

	for _, node := range g.Nodes {
		for _, dep := range node.Dependencies() {
			n, ok := d[dep.NodeID]
			if !ok {
				return nil, errors.New("dependency not found")
			}

			n.next = append(n.next, node.ID())
			d[dep.NodeID] = n
		}
	}

	return d, nil
}
