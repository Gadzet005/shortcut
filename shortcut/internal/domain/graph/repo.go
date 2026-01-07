package graph

type Repo interface {
	GetGraph(id ID) (Graph, error)
}
