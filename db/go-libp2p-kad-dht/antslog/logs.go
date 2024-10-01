package antslog

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
)

type RequestLog struct {
	Timestamp time.Time
	Self      peer.ID
	Requester peer.ID
	Type      uint8
	Target    mh.Multihash
	Maddrs    []ma.Multiaddr
	Agent     string
}
