package ants

import (
	"context"
	"fmt"

	ds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/ants"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/probe-lab/go-libdht/kad/key/bit256"
)

const (
	celestiaNet = Mainnet
	userAgent   = "celestiant"
)

type Ant struct {
	port    int
	dht     *kad.IpfsDHT
	privKey crypto.PrivKey

	Host      host.Host
	KadId     bit256.Key
	UserAgent string
}

func SpawnAnt(ctx context.Context, privKey crypto.PrivKey, peerstore peerstore.Peerstore, datastore ds.Batching, port int, logsChan chan ants.RequestEvent) (*Ant, error) {
	pid, _ := peer.IDFromPrivateKey(privKey)
	logger.Debugf("spawning ant. kadid: %s, peerid: %s", PeeridToKadid(pid).HexString(), pid)

	portStr := fmt.Sprint(port)

	// taken from github.com/celestiaorg/celestia-node/nodebuilder/p2p/config.go
	// ports are assigned automatically
	listenAddrs := []string{
		"/ip4/0.0.0.0/udp/" + portStr + "/quic-v1/webtransport",
		"/ip6/::/udp/" + portStr + "/quic-v1/webtransport",
		"/ip4/0.0.0.0/udp/" + portStr + "/quic-v1",
		"/ip6/::/udp/" + portStr + "/quic-v1",
		"/ip4/0.0.0.0/tcp/" + portStr,
		"/ip6/::/tcp/" + portStr,
	}

	opts := []libp2p.Option{
		libp2p.UserAgent(userAgent),
		libp2p.Identity(privKey),
		libp2p.Peerstore(peerstore),
		libp2p.DisableRelay(),

		libp2p.ListenAddrStrings(listenAddrs...),
	}

	if port == 0 {
		opts = append(opts, libp2p.NATPortMap()) // enable NAT port mapping if no port is specified
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		logger.Warn("unable to create libp2p host: ", err)
		return nil, err
	}

	dhtOpts := []kad.Option{
		kad.Mode(kad.ModeServer),
		kad.BootstrapPeers(BootstrapPeers(celestiaNet)...),
		kad.ProtocolPrefix(protocol.ID(fmt.Sprintf("/celestia/%s", celestiaNet))),
		kad.Datastore(datastore),
		kad.RequestsLogChan(logsChan),
	}
	dht, err := kad.New(ctx, h, dhtOpts...)
	if err != nil {
		logger.Warn("unable to create libp2p dht: ", err)
		return nil, err
	}

	ant := &Ant{
		Host:      h,
		dht:       dht,
		privKey:   privKey,
		KadId:     PeeridToKadid(h.ID()),
		port:      port,
		UserAgent: userAgent,
	}

	go dht.Bootstrap(ctx)

	return ant, nil
}

func (a *Ant) Close() error {
	err := a.dht.Close()
	if err != nil {
		return err
	}
	return a.Host.Close()
}
