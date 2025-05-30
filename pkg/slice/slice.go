package slice

func HasDuplicates(slice []string) bool {
	encountered := map[string]bool{}

	for _, v := range slice {
		if encountered[v] {
			return true
		}
		encountered[v] = true
	}

	return false
}

// Intersection returns things from b that are common and missing with a
func Intersection(a []string, b []string) ([]string, []string) {
	matches := make([]string, 0)
	misses := make([]string, 0)
	hash := make(map[string]bool)

	for _, aa := range a {
		hash[aa] = true
	}

	for _, bb := range b {
		if _, found := hash[bb]; found {
			matches = append(matches, bb)
		} else {
			misses = append(misses, bb)
		}
	}

	return matches, misses
}
