package ants

import (
	"context"
	"fmt"
	"time"

	ds "github.com/ipfs/go-datastore"
	p2pforge "github.com/ipshipyard/p2p-forge/client"
	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/ants"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	webtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"

	"github.com/probe-lab/go-libdht/kad/key/bit256"
)

const (
	celestiaNet = Mainnet
	userAgent   = "celestiant"
)

type AntConfig struct {
	PrivateKey     crypto.PrivKey
	UserAgent      string
	Port           int
	WssEnabled     bool
	WssPort        int
	ProtocolPrefix string
	BootstrapPeers []peer.AddrInfo
	EventsChan     chan ants.RequestEvent
}

func (cfg *AntConfig) Validate() error {
	if cfg.PrivateKey == nil {
		return fmt.Errorf("no ant private key given")
	}

	if cfg.UserAgent == "" {
		return fmt.Errorf("user agent is not set")
	}

	if cfg.ProtocolPrefix == "" {
		return fmt.Errorf("protocol prefix is not set")
	}

	if len(cfg.BootstrapPeers) == 0 {
		return fmt.Errorf("bootstrap peers are not set")
	}

	if cfg.EventsChan == nil {
		return fmt.Errorf("events channel is not set")
	}

	return nil
}

type Ant struct {
	cfg         *AntConfig
	host        host.Host
	dht         *kad.IpfsDHT
	kadID       bit256.Key
	certManager *p2pforge.P2PForgeCertMgr
}

func SpawnAnt(ctx context.Context, ps peerstore.Peerstore, ds ds.Batching, cfg *AntConfig) (*Ant, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no config given")
	} else if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	portStr := fmt.Sprint(cfg.Port)

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
		libp2p.Identity(cfg.PrivateKey),
		libp2p.Peerstore(ps),
		libp2p.DisableRelay(),
		libp2p.DisableMetrics(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
		libp2p.Transport(webtransport.New),
	}
	var certManager *p2pforge.P2PForgeCertMgr
	certLoaded := make(chan bool, 1)
	if cfg.WssEnabled {
		var err error
		certManager, err = p2pforge.NewP2PForgeCertMgr(
			p2pforge.WithOnCertLoaded(func() {
				certLoaded <- true
			}))
		if err != nil {
			logger.Warnf("wss cert manager error: %s", err)
		} else {
			listenAddrs = append(listenAddrs,
				fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/tls/sni/*.%s/ws", cfg.WssPort, p2pforge.DefaultForgeDomain),
				fmt.Sprintf("/ip6/::/tcp/%d/tls/sni/*.%s/ws", cfg.WssPort, p2pforge.DefaultForgeDomain),
			)
			opts = append(opts,
				libp2p.Transport(libp2pws.New, libp2pws.WithTLSConfig(certManager.TLSConfig())),
				libp2p.AddrsFactory(certManager.AddressFactory()),
			)
		}
	}

	opts = append(opts, libp2p.ListenAddrStrings(listenAddrs...))

	if cfg.Port == 0 {
		opts = append(opts, libp2p.NATPortMap()) // enable NAT port mapping if no port is specified
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	if certManager != nil {
		certManager.ProvideHost(h)
		if err := certManager.Start(); err != nil {
			logger.Warnf("wss cert manager start error: %s", err)
		}
		go func(certManager *p2pforge.P2PForgeCertMgr, certLoaded chan bool) {
			defer certManager.Stop()

			select {
			case <-certLoaded:
				logger.Info("certificate loaded: %s", h.ID())
			case <-time.After(time.Second * 70):
				logger.Warnf("timeout waiting for certificate: %s", h.ID())
			}
		}(certManager, certLoaded)
	}

	dhtOpts := []kad.Option{
		kad.Mode(kad.ModeServer),
		kad.BootstrapPeers(cfg.BootstrapPeers...),
		kad.ProtocolPrefix(protocol.ID(cfg.ProtocolPrefix)),
		kad.Datastore(ds),
		kad.RequestsLogChan(cfg.EventsChan),
	}
	dht, err := kad.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, fmt.Errorf("new libp2p dht: %w", err)
	}

	logger.Debugf("spawned ant. kadid: %s, peerid: %s", PeerIDToKadID(h.ID()).HexString(), h.ID())

	ant := &Ant{
		cfg:   cfg,
		host:  h,
		dht:   dht,
		kadID: PeerIDToKadID(h.ID()),
	}

	go dht.Bootstrap(ctx)

	return ant, nil
}

func (a *Ant) Close() error {
	err := a.dht.Close()
	if err != nil {
		return err
	}
	return a.host.Close()
}
