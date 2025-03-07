package ants

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
	mhreg "github.com/multiformats/go-multihash/core"

	"github.com/probe-lab/go-libdht/kad/key"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
	"github.com/probe-lab/go-libdht/kad/key/bitstr"
	"github.com/probe-lab/go-libdht/kad/trie"
)

type KeysDB struct {
	filepath string
}

func NewKeysDB(filepath string) *KeysDB {
	return &KeysDB{filepath: filepath}
}

func (db *KeysDB) readKeysFromFile() *trie.Trie[bit256.Key, crypto.PrivKey] {
	keysTrie := trie.New[bit256.Key, crypto.PrivKey]()

	// load file
	file, err := os.OpenFile(db.filepath, os.O_RDONLY, 0o600)
	if err != nil {
		logger.Warn("Couldn't open file", db.filepath, ":", err)
		return keysTrie
	}
	defer file.Close()

	for {
		// Read exactly ed25519.PrivateKeySize bytes into keyBytes
		keyBytes := make([]byte, ed25519.PrivateKeySize)
		_, err := io.ReadFull(file, keyBytes)
		if err == io.EOF {
			break // End of file, exit the loop
		}
		if err != nil {
			logger.Warnf("Error reading key: %v", err)
			break
		}

		// Unmarshal the private key
		privKey, err := crypto.UnmarshalEd25519PrivateKey(keyBytes)
		if err != nil {
			logger.Warnf("Error parsing key: %v", err)
			continue
		}
		// Derive the peer ID from the private key
		pid, err := peer.IDFromPrivateKey(privKey)
		if err != nil {
			logger.Warnf("Error getting peer ID: %v", err)
			continue
		}

		// Add to your keysTrie or equivalent data structure
		keysTrie.Add(PeerIDToKadID(pid), privKey)
	}
	return keysTrie
}

func (db *KeysDB) writeKeysToFile(keysTrie *trie.Trie[bit256.Key, crypto.PrivKey]) {
	file, err := os.OpenFile(db.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		logger.Warn("Couldn't open file", db.filepath, ":", err)
		return
	}
	defer file.Close()

	allKeys := trie.Closest(keysTrie, bit256.ZeroKey(), keysTrie.Size())
	for _, entry := range allKeys {
		if entry.Data == nil {
			fmt.Println(entry)
			continue
		}
		raw, err := entry.Data.Raw()
		if err != nil {
			logger.Warn("error getting raw key", err)
			continue
		}
		_, err = file.Write(raw)
		if err != nil {
			logger.Warn("error writing key to file", err)
		}
	}
}

// integrateKeysIntoTrie converts the provided privkeys into kademlia ids and
// adds them to the provided binary trie
func integrateKeysIntoTrie(keysTrie *trie.Trie[bit256.Key, crypto.PrivKey], keys []crypto.PrivKey) {
	for _, key := range keys {
		if key == nil {
			logger.Warn("skipping nil key")
			continue
		}

		pid, err := peer.IDFromPrivateKey(key)
		if err != nil {
			logger.Warnf("Error getting peer ID: %v", err)
			continue
		}
		keysTrie.Add(PeerIDToKadID(pid), key)
	}
}

func genKey() crypto.PrivKey {
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}
	return priv
}

// getMatchingKeys will return a list of private keys whose kademlia IDs match
// the provided list of prefixes, by looking for matches in the provided binary
// trie, and if no match by bruteforcing new keys until a match is found. All
// keys generated during bruteforces are added to the trie.
func getMatchingKeys(prefixes []bitstr.Key, keysTrie *trie.Trie[bit256.Key, crypto.PrivKey]) []crypto.PrivKey {
	// generate a random mask to be used as key suffix. If the same suffix is
	// used for all keys, the trie will be unbalanced
	randomMask := make([]byte, bit256.KeyLen)
	_, _ = rand.Read(randomMask)

	keys := make([]crypto.PrivKey, len(prefixes))
	// find or generate a key for each prefix
	for i, prefix := range prefixes {
		// apply random suffix to generate a 256-bit key from the prefix
		b256 := bitstrToBit256(prefix, randomMask)
		// find the closest key in the trie
		foundKeys := trie.Closest(keysTrie, b256, 1)
		if len(foundKeys) > 0 && key.CommonPrefixLength(foundKeys[0].Key, prefix) == prefix.BitLen() {
			// closest key is matching prefix
			keys[i] = foundKeys[0].Data
			// remove found key from db to avoid exposing private keys of live nodes
			keysTrie.Remove(foundKeys[0].Key)
		} else {
			// no matching key found in trie, generate new keys until one matches
			for {
				newKey := genKey()
				// derive peer ID from the private key
				pid, err := peer.IDFromPrivateKey(newKey)
				if err != nil {
					logger.Warnf("Error getting peer ID: %v", err)
					continue
				}
				// check if the new key matches the prefix
				if key.CommonPrefixLength(PeerIDToKadID(pid), prefix) == prefix.BitLen() {
					keys[i] = newKey
					break
				}
				// add to keysTrie if not matching prefix
				keysTrie.Add(PeerIDToKadID(pid), newKey)
			}
		}
	}
	return keys
}

// MatchingKeys returns a list of private keys whose kademlia IDs match the
// provided list of prefixes. It also write back to disk the returned private
// keys for future use.
func (db *KeysDB) MatchingKeys(prefixes []bitstr.Key, returned []crypto.PrivKey) []crypto.PrivKey {
	// read keys from disk
	keysTrie := db.readKeysFromFile()

	// integrate returned keys into the trie, they are available again
	integrateKeysIntoTrie(keysTrie, returned)

	// pop any matching keys from the keysTrie, generate new keys if needed
	// and store them in keysTrie
	keys := getMatchingKeys(prefixes, keysTrie)

	// save keys to disk
	db.writeKeysToFile(keysTrie)

	return keys
}

// PeerIDToKadID converts a libp2p peer.ID to its binary kademlia identifier
func PeerIDToKadID(pid peer.ID) bit256.Key {
	hasher, err := mhreg.GetHasher(mh.SHA2_256)
	if err != nil {
		panic(err)
	}
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
