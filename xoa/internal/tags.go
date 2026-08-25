package internal

import "slices"

// TagsMatch returns true when every wanted tag is present in have.
// It implements the same all-of semantics the SDK uses to compare tags
// in its Compare methods.
func TagsMatch(wanted, have []string) bool {
	for _, t := range wanted {
		if !slices.Contains(have, t) {
			return false
		}
	}
	return true
}
