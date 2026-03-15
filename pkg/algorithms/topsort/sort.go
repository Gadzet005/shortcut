package topsort

// Sort возвращает топологически отсортированные уровни графа.
// Если граф содержит цикл, возвращает nil и false, иначе возвращает уровни и true.
func Sort(g map[string][]string) ([][]string, bool) {
	if len(g) == 0 {
		return nil, true
	}

	allNodes := make(map[string]struct{})
	for nodeID := range g {
		allNodes[nodeID] = struct{}{}
	}
	for _, neighbors := range g {
		for _, id := range neighbors {
			allNodes[id] = struct{}{}
		}
	}

	inDegree := make(map[string]int)
	for nodeID := range allNodes {
		inDegree[nodeID] = 0
	}
	for _, neighbors := range g {
		for _, nextID := range neighbors {
			inDegree[nextID]++
		}
	}

	var currentLevel []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			currentLevel = append(currentLevel, nodeID)
		}
	}

	if len(currentLevel) == 0 {
		return nil, false
	}

	var result [][]string
	visited := 0

	for len(currentLevel) > 0 {
		result = append(result, currentLevel)
		visited += len(currentLevel)

		var nextLevel []string
		nextLevelSet := make(map[string]bool)

		for _, nodeID := range currentLevel {
			for _, nextID := range g[nodeID] {
				inDegree[nextID]--
				if inDegree[nextID] == 0 && !nextLevelSet[nextID] {
					nextLevel = append(nextLevel, nextID)
					nextLevelSet[nextID] = true
				}
			}
		}

		currentLevel = nextLevel
	}

	if visited != len(allNodes) {
		return nil, false
	}

	return result, true
}
