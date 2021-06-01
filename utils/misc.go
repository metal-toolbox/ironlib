package utils

// StringInSlice returns a bool value if the given string was not/found in the given slice.
func StringInSlice(str string, sl []string) bool {
	for _, element := range sl {
		if element == str {
			return true
		}
	}

	return false
}
