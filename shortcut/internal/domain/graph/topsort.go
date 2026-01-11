package graph

import "errors"

func TopSort(g Graph) ([][]NodeID, error) {
	d, err := convertToDAG(g)
	if err != nil {
		return nil, err
	}
	return topSortDag(d)
}

func topSortDag(d dag) ([][]NodeID, error) {
	inDegree := make(map[NodeID]int)
	for nodeID := range d {
		inDegree[nodeID] = 0
	}

	for _, node := range d {
		for _, nextID := range node.next {
			inDegree[nextID]++
		}
	}

	var currentLevel []NodeID
	for nodeID, degree := range inDegree {
		if degree == 0 {
			currentLevel = append(currentLevel, nodeID)
		}
	}

	if len(currentLevel) == 0 {
		return nil, errors.New("graph has a cycle or is empty")
	}

	var result [][]NodeID
	visited := 0

	for len(currentLevel) > 0 {
		result = append(result, currentLevel)
		visited += len(currentLevel)

		var nextLevel []NodeID
		nextLevelSet := make(map[NodeID]bool)

		for _, nodeID := range currentLevel {
			node := d[nodeID]
			for _, nextID := range node.next {
				inDegree[nextID]--
				if inDegree[nextID] == 0 && !nextLevelSet[nextID] {
					nextLevel = append(nextLevel, nextID)
					nextLevelSet[nextID] = true
				}
			}
		}

		currentLevel = nextLevel
	}

	if visited != len(d) {
		return nil, errors.New("graph has a cycle")
	}

	return result, nil
}
