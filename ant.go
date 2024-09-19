package ants

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	antslog "github.com/libp2p/go-libp2p-kad-dht/antslog"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/probe-lab/go-libdht/kad/key/bit256"
)

const (
	celestiaNet = Mainnet
)

type Ant struct {
	Host host.Host
	dht  *kad.IpfsDHT

	KadId bit256.Key
}

func SpawnAnt(ctx context.Context, privKey crypto.PrivKey, logsChan chan antslog.RequestLog) (*Ant, error) {
	pid, _ := peer.IDFromPrivateKey(privKey)
	logger.Debugf("spawning ant. kadid: %s, peerid: %s", PeeridToKadid(pid).HexString(), pid)

	// taken from github.com/celestiaorg/celestia-node/nodebuilder/p2p/config.go
	// ports are assigned automatically
	listenAddrs := []string{
		"/ip4/0.0.0.0/udp/0/quic-v1/webtransport",
		"/ip6/::/udp/0/quic-v1/webtransport",
		"/ip4/0.0.0.0/udp/0/quic-v1",
		"/ip6/::/udp/0/quic-v1",
		"/ip4/0.0.0.0/tcp/0",
		"/ip6/::/tcp/0",
	}

	opts := []libp2p.Option{
		libp2p.UserAgent("celestiant"),
		libp2p.Identity(privKey),
		libp2p.NATPortMap(), // enable uPnP
		libp2p.DisableRelay(),

		libp2p.ListenAddrStrings(listenAddrs...),
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
		kad.RequestsLogChan(logsChan),
	}
	dht, err := kad.New(ctx, h, dhtOpts...)
	if err != nil {
		logger.Warn("unable to create libp2p dht: ", err)
		return nil, err
	}

	ant := &Ant{
		Host:  h,
		dht:   dht,
		KadId: PeeridToKadid(h.ID()),
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