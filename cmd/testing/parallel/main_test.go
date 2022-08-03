package main

import (
	"os"
	"testing"
	"time"
)

var c chan bool

func TestMain(m *testing.M) {
	c = make(chan bool, 1)

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(10 * time.Second)
			c <- true
		}
	}()

	os.Exit(m.Run())
}

func waitForChannel(channel chan bool) {
	<-channel
}

func TestParallelSuccessful(t *testing.T) {
	t.Parallel()
	waitForChannel(c)
}

func TestParallelHangging(t *testing.T) {
	t.Parallel()
	waitForChannel(c)
}

func TestParallelSlightHangging(t *testing.T) {
	t.Parallel()
	waitForChannel(c)
}
