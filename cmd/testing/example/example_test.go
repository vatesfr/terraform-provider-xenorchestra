// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package example

import (
	"testing"
	"time"
)

func TestNextStartWithOptions(t *testing.T) {
	time.Sleep(1000 * time.Millisecond)
}
func TestDummy(t *testing.T) {
	time.Sleep(1000 * time.Millisecond)
}

func TestStart(t *testing.T) {
	time.Sleep(1500 * time.Millisecond)
}

var runs = 0

func TestFlakyTest(t *testing.T) {
	time.Sleep(300 * time.Millisecond)
	if runs < 1 {
		runs += 1
		t.Fail()
	}
}

func TestLoading(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
}

func TestLoading_abort(t *testing.T) {
	t.Parallel()
	time.Sleep(2500 * time.Millisecond)
}

func TestLoading_interrupt(t *testing.T) {
	t.Parallel()
	time.Sleep(80 * time.Millisecond)
}
