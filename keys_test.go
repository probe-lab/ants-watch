package ants

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/probe-lab/go-libdht/kad/key"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/probe-lab/go-libdht/kad/trie"
	"github.com/stretchr/testify/require"
)

func TestWriteReadKeys(t *testing.T) {
	nKeys := 128
	filename := "test_keys.db"

	keysTrie := trie.New[bit256.Key, crypto.PrivKey]()
	for i := 0; i < nKeys; i++ {
		priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
		require.NoError(t, err)
		pid, err := peer.IDFromPrivateKey(priv)
		require.NoError(t, err)
		keysTrie.Add(PeerIDToKadID(pid), priv)
	}

	db := NewKeysDB(filename)
	db.writeKeysToFile(keysTrie)

	resultTrie := db.readKeysFromFile()
	require.Equal(t, keysTrie.Size(), resultTrie.Size())

	// verify that all keys are present
	for _, k := range trie.Closest(keysTrie, bit256.ZeroKey(), keysTrie.Size()) {
		present, privKey := trie.Find(resultTrie, k.Key)
		require.True(t, present)
		require.Equal(t, k.Data, privKey)
	}

	os.Remove(filename)
}

func genPrefixesForDepth(depth int) []bitstr.Key {
	prefixes := []string{""}
	for i := 0; i < depth; i++ {
		newPrefixes := []string{}
		for _, prefix := range prefixes {
			newPrefixes = append(newPrefixes, prefix+"0", prefix+"1")
		}
		prefixes = newPrefixes
	}
	prefixKeys := make([]bitstr.Key, len(prefixes))
	for i, prefix := range prefixes {
		prefixKeys[i] = bitstr.Key(prefix)
	}
	return prefixKeys
}

func TestKeysDB(t *testing.T) {
	filename := "test_keys.db"
	prefixDepth := 4 // 2**(prefixDepth) prefixes

	defer os.Remove(filename)

	prefixes := genPrefixesForDepth(prefixDepth)

	db := NewKeysDB(filename)
	require.NotNil(t, db)

	privKeys := db.MatchingKeys(prefixes, nil)

	// check that the returned private keys match the prefixes
	require.Len(t, privKeys, len(prefixes))
	for i, prefix := range prefixes {
		pid, err := peer.IDFromPrivateKey(privKeys[i])
		require.NoError(t, err)
		require.Equal(t, prefix.BitLen(), key.CommonPrefixLength(PeerIDToKadID(pid), prefix))
	}

	// check that the keys are not reused
	privKeys2 := db.MatchingKeys(prefixes, nil)
	require.Len(t, privKeys2, len(prefixes))
	for _, key := range privKeys {
		require.NotContains(t, privKeys2, key)
	}
}

func TestReturnKeysToEmptyTrie(t *testing.T) {
	filename := "test_keys.db"
	db := NewKeysDB(filename)
	defer os.Remove(filename)

	key := genKey()
	privKeys := db.MatchingKeys(nil, []crypto.PrivKey{key})
	require.Len(t, privKeys, 0)

	keysTrie := db.readKeysFromFile()
	require.Equal(t, 1, keysTrie.Size())

	pid, err := peer.IDFromPrivateKey(key)
	require.NoError(t, err)
	found, foundKey := trie.Find(keysTrie, PeerIDToKadID(pid))
	require.True(t, found)
	require.Equal(t, key, foundKey)
}
