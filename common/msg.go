package common

import "time"

type HelloArgs struct {
	Msg string
}

type HelloReply string

const (
	// Maximum message size allowed from peer.
	MaxMessageSize = 4096
	PongWait       = 60 * time.Second
)
