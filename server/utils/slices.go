package utils

// SlicesContains reports whether v is present in s.
// https://pkg.go.dev/golang.org/x/exp/slices#Contains
func SlicesContains[E comparable](s []E, v E) bool {
	for _, vs := range s {
		if v == vs {
			return true
		}
	}
	return false
}
