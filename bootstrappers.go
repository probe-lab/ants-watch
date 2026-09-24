package ants

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// NetworkID identifies a network by its project and sub-network, e.g.
// {Project: "celestia", Network: "mainnet"}. It is both what the queen requests
// from the nebula service and what selects the DHT parameters below.
type NetworkID struct {
	Project string
	Network string
}

func (n NetworkID) String() string {
	return n.Project + "/" + n.Network
}

// NetworkConfig holds the DHT parameters for a single network.
type NetworkConfig struct {
	ProtocolID     string
	UserAgent      string
	BootstrapPeers []peer.AddrInfo
}

// NOTE: Every time we add a new long-running network, add its parameters here.
var networks = map[NetworkID]struct {
	protocolID     string
	userAgent      string
	bootstrapPeers []string
}{
	{Project: "celestia", Network: "mainnet"}: {
		protocolID: "/celestia/celestia/kad/1.0.0",
		userAgent:  "probelab-node/celestia/ant/v0.1.0",
		bootstrapPeers: []string{
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
	},
	{Project: "celestia", Network: "arabica-11"}: {
		protocolID: "/celestia/arabica-11/kad/1.0.0",
		userAgent:  "probelab-node/celestia/ant/v0.1.0",
		bootstrapPeers: []string{
			"/dnsaddr/da-bridge-1.celestia-arabica-11.com/p2p/12D3KooWGqwzdEqM54Dce6LXzfFr97Bnhvm6rN7KM7MFwdomfm4S",
			"/dnsaddr/da-bridge-2.celestia-arabica-11.com/p2p/12D3KooWCMGM5eZWVfCN9ZLAViGfLUWAfXP5pCm78NFKb9jpBtua",
			"/dnsaddr/da-bridge-3.celestia-arabica-11.com/p2p/12D3KooWEWuqrjULANpukDFGVoHW3RoeUU53Ec9t9v5cwW3MkVdQ",
			"/dnsaddr/da-bridge-4.celestia-arabica-11.com/p2p/12D3KooWLT1ysSrD7XWSBjh7tU1HQanF5M64dHV6AuM6cYEJxMPk",
		},
	},
	{Project: "celestia", Network: "mocha-4"}: {
		protocolID: "/celestia/mocha-4/kad/1.0.0",
		userAgent:  "probelab-node/celestia/ant/v0.1.0",
		bootstrapPeers: []string{
			"/dns4/da-bridge-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWCBAbQbJSpCpCGKzqz3rAN4ixYbc63K68zJg9aisuAajg",
			"/dns4/da-bridge-mocha-4-2.celestia-mocha.com/tcp/2121/p2p/12D3KooWK6wJkScGQniymdWtBwBuU36n6BRXp9rCDDUD6P5gJr3G",
			"/dns4/da-full-1-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWCUHPLqQXZzpTx1x3TAsdn3vYmTNDhzg66yG8hqoxGGN8",
			"/dns4/da-full-2-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWR6SHsXPkkvhCRn6vp1RqSefgaT1X1nMNvrVjU2o3GoYy",
		},
	},
	{Project: "avail", Network: "mainnet-lc"}: {
		protocolID: "/avail_kad/id/1.0.0-b91746",
		// Spoof agent version because of this check:
		// https://github.com/availproject/avail-light/blob/2bd85abd4eb502c818e3cd634bd235fea477571f/core/src/network/p2p/event_loop.rs#L441
		userAgent: "avail-light-client/light-client/1.12.13/go-ant",
		bootstrapPeers: []string{
			"/dns/bootnode.1.lightclient.mainnet.avail.so/tcp/37000/p2p/12D3KooW9x9qnoXhkHAjdNFu92kMvBRSiFBMAoC5NnifgzXjsuiM",
		},
	},
	{Project: "agntcy", Network: "mainnet"}: {
		// no leading slash: agntcy's ProtocolPrefix is "dir"
		protocolID: "dir/kad/1.0.0",
		userAgent:  "probelab-node/agntcy/ant/v0.1.0",
		bootstrapPeers: []string{
			"/dns4/routing.ads.outshift.io/tcp/5555/p2p/12D3KooWLf9p3cedc86xGQBaqak6rAFmQk1HxKAK1yh7umHE3amu",
		},
	},
}

// LookupNetwork returns the DHT configuration for the given project and network.
// The boolean is false when the network is unknown or has no configured
// bootstrap peers - in which case the queen must not start, since its ants
// could not join the DHT.
func LookupNetwork(project, network string) (NetworkConfig, bool) {
	raw, ok := networks[NetworkID{Project: project, Network: network}]
	if !ok || len(raw.bootstrapPeers) == 0 {
		return NetworkConfig{}, false
	}

	peers := make([]peer.AddrInfo, len(raw.bootstrapPeers))
	for i, addr := range raw.bootstrapPeers {
		ai, err := peer.AddrInfoFromString(addr)
		if err != nil {
			panic(err)
		}
		peers[i] = *ai
	}

	return NetworkConfig{
		ProtocolID:     raw.protocolID,
		UserAgent:      raw.userAgent,
		BootstrapPeers: peers,
	}, true
}
