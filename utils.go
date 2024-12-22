package main

import "regexp"

// IsNumeric checks if a string contains only numeric characters.
func IsNumeric(s string) bool {
	// Compile the regex pattern
	re := regexp.MustCompile(`^[0-9]+$`)
	return re.MatchString(s)
}
