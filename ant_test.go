package ants

import (
	"testing"
	"time"

	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

func testRequestEvent(t *testing.T) RequestEvent {
	t.Helper()
	target, err := mh.Sum([]byte("record"), mh.SHA2_256, -1)
	require.NoError(t, err)
	return RequestEvent{
		Timestamp: time.Now(),
		Self:      peer.ID("ant"),
		Remote:    peer.ID("remote"),
		Type:      pb.Message_ADD_PROVIDER,
		Target:    target,
	}
}

func TestRequestEventHash(t *testing.T) {
	base := testRequestEvent(t)

	same := base
	same.Timestamp = base.Timestamp.Add(time.Hour)
	same.Self = peer.ID("other-ant")
	require.Equal(t, base.Hash(), same.Hash())

	otherPeer := base
	otherPeer.Remote = peer.ID("remote-2")
	require.NotEqual(t, base.Hash(), otherPeer.Hash())

	otherType := base
	otherType.Type = pb.Message_FIND_NODE
	require.NotEqual(t, base.Hash(), otherType.Hash())

	otherKey := base
	otherKey.Target = append(mh.Multihash{}, base.Target...)
	otherKey.Target[len(otherKey.Target)-1] ^= 0xff
	require.NotEqual(t, base.Hash(), otherKey.Hash())
}
