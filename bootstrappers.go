package ants

import "github.com/libp2p/go-libp2p/core/peer"

type Network string

const (
	// DefaultNetwork is the default network of the current build.
	DefaultNetwork = Mainnet
	// Arabica testnet. See: celestiaorg/networks.
	Arabica Network = "arabica-11"
	// Mocha testnet. See: celestiaorg/networks.
	Mocha Network = "mocha-4"
	// Private can be used to set up any private network, including local testing setups.
	Private Network = "private"
	// Celestia mainnet. See: celestiaorg/networks.
	Mainnet Network = "celestia"
)

// NOTE: Every time we add a new long-running network, its bootstrap peers have to be added here.
var bootstrapList = map[Network][]string{
	Mainnet: {
		"/dns4/da-bridge-1.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWSqZaLcn5Guypo2mrHr297YPJnV8KMEMXNjs3qAS8msw8",
		"/dns4/da-bridge-2.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWQpuTFELgsUypqp9N4a1rKBccmrmQVY8Em9yhqppTJcXf",
		"/dns4/da-bridge-3.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWSGa4huD6ts816navn7KFYiStBiy5LrBQH1HuEahk4TzQ",
		"/dns4/da-bridge-4.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWHBXCmXaUNat6ooynXG837JXPsZpSTeSzZx6DpgNatMmR",
		"/dns4/da-bridge-5.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWDGTBK1a2Ru1qmnnRwP6Dmc44Zpsxi3xbgFk7ATEPfmEU",
		"/dns4/da-bridge-6.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWLTUFyf3QEGqYkHWQS2yCtuUcL78vnKBdXU5gABM1YDeH",
		"/dns4/da-full-1.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWKZCMcwGCYbL18iuw3YVpAZoyb1VBGbx9Kapsjw3soZgr",
		"/dns4/da-full-2.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWE3fmRtHgfk9DCuQFfY3H3JYEnTU3xZozv1Xmo8KWrWbK",
		"/dns4/da-full-3.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWK6Ftsd4XsWCsQZgZPNhTrE5urwmkoo5P61tGvnKmNVyv",
	},
	Arabica: {
		"/dnsaddr/da-bridge-1.celestia-arabica-11.com/p2p/12D3KooWGqwzdEqM54Dce6LXzfFr97Bnhvm6rN7KM7MFwdomfm4S",
		"/dnsaddr/da-bridge-2.celestia-arabica-11.com/p2p/12D3KooWCMGM5eZWVfCN9ZLAViGfLUWAfXP5pCm78NFKb9jpBtua",
		"/dnsaddr/da-bridge-3.celestia-arabica-11.com/p2p/12D3KooWEWuqrjULANpukDFGVoHW3RoeUU53Ec9t9v5cwW3MkVdQ",
		"/dnsaddr/da-bridge-4.celestia-arabica-11.com/p2p/12D3KooWLT1ysSrD7XWSBjh7tU1HQanF5M64dHV6AuM6cYEJxMPk",
	},
	Mocha: {
		"/dns4/da-bridge-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWCBAbQbJSpCpCGKzqz3rAN4ixYbc63K68zJg9aisuAajg",
		"/dns4/da-bridge-mocha-4-2.celestia-mocha.com/tcp/2121/p2p/12D3KooWK6wJkScGQniymdWtBwBuU36n6BRXp9rCDDUD6P5gJr3G",
		"/dns4/da-full-1-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWCUHPLqQXZzpTx1x3TAsdn3vYmTNDhzg66yG8hqoxGGN8",
		"/dns4/da-full-2-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWR6SHsXPkkvhCRn6vp1RqSefgaT1X1nMNvrVjU2o3GoYy",
	},
	Private: {},
}

func BootstrapPeers(net Network) []peer.AddrInfo {
	peers := make([]peer.AddrInfo, len(bootstrapList[net]))
	for i, addr := range bootstrapList[net] {
		peer, err := peer.AddrInfoFromString(addr)
		if err != nil {
			panic(err)
		}
		peers[i] = *peer
	}
	return peers
}
