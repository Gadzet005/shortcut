package sets

func New[T comparable]() *Set[T] {
	return &Set[T]{m: make(map[T]bool)}
}

func NewFromSlice[T comparable](arr []T) *Set[T] {
	ret := &Set[T]{m: make(map[T]bool)}

	for _, v := range(arr) {
		ret.Add(v)
	}

	return ret
}

type Set[T comparable] struct {
	m map[T]bool
}

func (s *Set[T]) Add(t T) {
	s.m[t] = true
}

func (s *Set[T]) Erase(t T) {
	delete(s.m, t)
}

func (s *Set[T]) Contains(t T) bool {
	_, ok := s.m[t]
	return ok
}

func (s *Set[T]) Size() int {
	return len(s.m)
}

func (s *Set[T]) Clear() {
	s.m = make(map[T]bool)
}

func (s *Set[T]) AsSlice() []T {
	ret := make([]T, 0)

	for k := range(s.m) {
		ret = append(ret, k)
	}

	return ret
}
