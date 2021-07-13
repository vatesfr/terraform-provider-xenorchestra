package internal

import (
	"bytes"
	"fmt"
	"hash/crc32"
)

// This was created due to terraform-plugin-sdk removing it in v2

// String hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func String(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// Strings hashes a list of strings to a unique hashcode.
func Strings(strings []string) string {
	var buf bytes.Buffer

	for i, s := range strings {
		var format = "%s"
		if i != len(strings)-1 {
			format = fmt.Sprint(format, "-")
		}
		buf.WriteString(fmt.Sprintf(format, s))
	}

	return fmt.Sprintf("%d", String(buf.String()))
}
