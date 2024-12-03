package ants

import (
	"testing"

	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/stretchr/testify/require"
)

func TestBitstrToBit256(t *testing.T) {
	strKey := bitstr.Key("00001111") // 0x0f
	var padding [bit256.KeyLen]byte
	got := bitstrToBit256(strKey, padding[:])

	var target [bit256.KeyLen]byte
	target[0] = 0x0f
	want := bit256.NewKey(target[:])
	require.Equal(t, want, got)

	strKey = "000011110000" // 0x0f0
	got = bitstrToBit256(strKey, padding[:])
	require.Equal(t, want, got)

	strKey = "111" // 0xe
	padding[0] = 0x0f
	got = bitstrToBit256(strKey, padding[:])
	target[0] = 0xef
	want = bit256.NewKey(target[:])
	require.Equal(t, want, got)
}
