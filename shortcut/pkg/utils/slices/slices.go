package slices

import "golang.org/x/exp/constraints"

type Number interface {
    constraints.Float | constraints.Integer
}

func Map[T, S any](arr []T, f func(T) S) []S {
	ret := make([]S, len(arr))

	for i, t := range(arr) {
		ret[i] = f(t)
	}

	return ret
}

func Contains[T comparable](arr []T, val T) bool {
	for _, v := range(arr) {
		if v == val {
			return true
		}
	}

	return false
}

func Sum[T Number](arr []T) T {
	var total T

	for _, n := range(arr) {
		total += n
	}

	return total
}
