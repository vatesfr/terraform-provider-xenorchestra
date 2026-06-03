package internal

import (
	"sort"
	"strings"

	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

// SortPools sorts a list of pools based on the specified field and order
func SortPools(pools []client.Pool, sortBy, sortOrder string) []client.Pool {
	if len(pools) == 0 {
		return pools
	}

	switch sortBy {
	case "id":
		sort.Slice(pools, func(i, j int) bool {
			return compareString(pools[i].Id, pools[j].Id, sortOrder)
		})
	case "name_label":
		sort.Slice(pools, func(i, j int) bool {
			return compareString(pools[i].NameLabel, pools[j].NameLabel, sortOrder)
		})
	default:
		// No sorting if sort_by is not recognized
		return pools
	}

	return pools
}

// compareString compares two strings based on the sort order
// Returns true if a should come before b
func compareString(a, b string, sortOrder string) bool {
	if sortOrder == "desc" {
		return strings.ToLower(b) < strings.ToLower(a)
	}
	// Default to "asc"
	return strings.ToLower(a) < strings.ToLower(b)
}
