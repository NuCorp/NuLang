package slicesutil

import "github.com/LicorneSharing/GTL/iter"

func Map[E, T any, S ~[]E](slice S, f func(E) T) []T {
	ret := make([]T, len(slice))

	for i, v := range slice {
		ret[i] = f(v)
	}

	return ret
}

func MapSeq[E, T any](seq iter.Seq[E], f func(E) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		seq(
			func(e E) bool {
				return yield(f(e))
			},
		)
	}
}
