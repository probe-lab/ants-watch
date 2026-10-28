package ants

import (
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/stretchr/testify/require"
)

func TestQueenIsDuplicate(t *testing.T) {
	window := 50 * time.Millisecond
	q := &Queen{seenRPCs: expirable.NewLRU[[32]byte, struct{}](100, nil, window)}
	evt := testRequestEvent(t)

	require.False(t, q.isDuplicate(evt))
	require.True(t, q.isDuplicate(evt))

	findNode := evt
	findNode.Type = pb.Message_FIND_NODE
	require.False(t, q.isDuplicate(findNode))

	time.Sleep(2 * window)
	require.False(t, q.isDuplicate(evt))
}

func TestQueenIsDuplicateDisabled(t *testing.T) {
	q := &Queen{}
	evt := testRequestEvent(t)

	require.False(t, q.isDuplicate(evt))
	require.False(t, q.isDuplicate(evt))
}
