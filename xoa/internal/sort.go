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

// SortStorageRepositories sorts a list of storage repositories based on the
// specified field and order
func SortStorageRepositories(srs []client.StorageRepository, sortBy, sortOrder string) []client.StorageRepository {
	if len(srs) == 0 {
		return srs
	}

	switch sortBy {
	case "id":
		sort.Slice(srs, func(i, j int) bool {
			return compareString(srs[i].Id, srs[j].Id, sortOrder)
		})
	case "name_label":
		sort.Slice(srs, func(i, j int) bool {
			return compareString(srs[i].NameLabel, srs[j].NameLabel, sortOrder)
		})
	case "size":
		sort.Slice(srs, func(i, j int) bool {
			return compareInt(srs[i].Size, srs[j].Size, sortOrder)
		})
	case "physical_usage":
		sort.Slice(srs, func(i, j int) bool {
			return compareInt(srs[i].PhysicalUsage, srs[j].PhysicalUsage, sortOrder)
		})
	case "usage":
		sort.Slice(srs, func(i, j int) bool {
			return compareInt(srs[i].Usage, srs[j].Usage, sortOrder)
		})
	default:
		// No sorting if sort_by is not recognized
		return srs
	}

	return srs
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

// compareInt compares two ints based on the sort order
// Returns true if a should come before b
func compareInt(a, b int, sortOrder string) bool {
	if sortOrder == "desc" {
		return b < a
	}
	// Default to "asc"
	return a < b
}
