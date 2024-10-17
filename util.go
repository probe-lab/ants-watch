package ants

import (
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"

	mh "github.com/multiformats/go-multihash"
	mhreg "github.com/multiformats/go-multihash/core"

	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
)

const (
	CRAWL_INTERVAL         = 30 * time.Minute
	NORMALIZATION_INTERVAL = 60 * time.Second
	BUCKET_SIZE            = 20
)

func PeeridToKadid(pid peer.ID) bit256.Key {
	hasher, _ := mhreg.GetHasher(mh.SHA2_256)
	hasher.Write([]byte(pid))
	return bit256.NewKey(hasher.Sum(nil))
}

func NsToCid(ns string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(ns), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}

	return cid.NewCidV1(cid.Raw, h), nil
}

func bitstrToBit256(strKey bitstr.Key, padding []byte) bit256.Key {
	bit256Key := make([]byte, 32)
	copy(bit256Key, padding)

	var currByte byte
	i := 0
	for ; i < strKey.BitLen(); i++ {
		currByte = currByte | (byte(strKey.Bit(i)) << (7 - (i % 8)))
		if i%8 == 7 {
			bit256Key[i/8] = currByte
			currByte = 0
		}
	}
	if i%8 != 0 {
		currByte = currByte | bit256Key[i/8]<<(7-(i%8))>>(7-(i%8))
		bit256Key[i/8] = currByte
	}
	return bit256.NewKey(bit256Key)
}
