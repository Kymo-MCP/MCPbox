package utils

// RemoveDuplicates
func RemoveDuplicates[T comparable](slice []T) []T {
	encountered := make(map[T]bool)
	result := make([]T, 0)

	for _, item := range slice {
		if !encountered[item] {
			encountered[item] = true
			result = append(result, item)
		}
	}

	return result
}
