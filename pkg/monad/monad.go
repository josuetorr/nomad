package monad

func Filter[T any](s []T, p func(T) bool) []T {
	res := []T{}
	for _, t := range s {
		if !p(t) {
			continue
		}
		res = append(res, t)
	}
	return res
}

func Chopn[T any](s []T, n int) []T {
	res := []T{}
	for i, t := range s {
		if (i + 1) == n {
			return res
		}
		res = append(res, t)
	}
	return res
}
