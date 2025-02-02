package lib

import "unique"

// RemoveDuplicates returns a new slice with duplicate strings removed.
func RemoveDuplicates(input []string) []string {
	seen := make(map[unique.Handle[string]]bool)
	var result []string

	for _, s := range input {
		h := unique.Make(s)
		if !seen[h] {
			seen[h] = true
			result = append(result, h.Value())
		}
	}
	return result
}

// RemoveEmptyString return a new slice without empty string
func RemoveEmptyString(input []string) []string {
	result := make([]string, 0, len(input))
	for _, s := range input {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
