package xoa

import "testing"

func TestWithout(t *testing.T) {
	old := ListTags([]string{"one", "two", "three"})
	new := ListTags([]string{"two", "three"})

	tags := old.Without(new)

	l := len(tags)
	if l != 1 {
		t.Fatalf("expected only one tag to remain, instead found '%d'", l)
	}

	if tags[0] != "one" {
		t.Errorf("expected the remaining tags to include 'one', instead found '%v'", tags)
	}
}
