package util

func Contains[T comparable](haystack []T, needle T) bool {
	for _, element := range haystack {
		if element == needle {
			return true
		}
	}

	return false
}
