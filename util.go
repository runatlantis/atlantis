package main

// ContainsStr returns true if slice has a value equal to search and false otherwise
func ContainsStr(slice []string, search string) bool {
	for _, v := range slice {
		if v == search {
			return true
		}
	}
	return false
}
