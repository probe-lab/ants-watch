package ants

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/probe-lab/go-libdht/kad/key/bit256"
)

const (
	celestiaNet = Mainnet
)

type Ant struct {
	host host.Host
	dht  *kad.IpfsDHT

	closeChan chan struct{}

	KadId bit256.Key
}

func SpawnAnt(ctx context.Context, privKey crypto.PrivKey) (*Ant, error) {
	// TODO: edit libp2p host for cloud deployment
	h, err := libp2p.New(
		libp2p.UserAgent("celestia-celestia"),
		libp2p.Identity(privKey),
		libp2p.NATPortMap(),
		libp2p.DisableRelay(),
	)
	if err != nil {
		return nil, err
	}

	dhtOpts := []kad.Option{
		kad.Mode(kad.ModeServer),
		kad.BootstrapPeers(BootstrapPeers(celestiaNet)...),
		kad.ProtocolPrefix(protocol.ID(fmt.Sprintf("/celestia/%s", celestiaNet))),
	}
	dht, err := kad.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, err
	}

	ant := &Ant{
		host:      h,
		dht:       dht,
		closeChan: make(chan struct{}, 1),
		KadId:     PeeridToKadid(h.ID()),
	}

	ant.run(ctx)

	return ant, nil
}

func (a *Ant) run(ctx context.Context) {
	a.dht.Bootstrap(ctx)

	// TODO: log events
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.closeChan:
			return
		}
	}
}

// TODO: double check if this is the correct way to close the ant
func (a *Ant) Close() error {
	a.closeChan <- struct{}{}
	err := a.dht.Close()
	if err != nil {
		return err
	}
	return a.host.Close()
}
