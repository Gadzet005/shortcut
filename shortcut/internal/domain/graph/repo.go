package graph

type GraphRepo interface {
	GetGraph(id GraphID) (Graph, error)
}
