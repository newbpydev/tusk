package app

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b { // Corrected logic: return a if a > b
		return a
	}
	return b
}

// Note: Add any other general utility functions here as needed.
// For example, date parsing/formatting helpers could live here if used in multiple places.
