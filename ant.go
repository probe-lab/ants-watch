package ants

import (
	"context"
	"fmt"

	ds "github.com/ipfs/go-datastore"
	p2pforge "github.com/ipshipyard/p2p-forge/client"
	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/ants"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pwebrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	libp2pws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	libp2pwebtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"

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
	PortWSS        int
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
	cfg            *AntConfig
	host           host.Host
	dht            *kad.IpfsDHT
	certMgr        *p2pforge.P2PForgeCertMgr
	kadID          bit256.Key
	certLoadedChan chan struct{}
	sub            event.Subscription
}

func SpawnAnt(ctx context.Context, ps peerstore.Peerstore, ds ds.Batching, cfg *AntConfig) (*Ant, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no config given")
	} else if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	certLoadedChan := make(chan struct{})
	forgeDomain := p2pforge.DefaultForgeDomain
	certMgr, err := p2pforge.NewP2PForgeCertMgr(
		p2pforge.WithForgeDomain(forgeDomain),
		p2pforge.WithOnCertLoaded(func() {
			certLoadedChan <- struct{}{}
		}),
		p2pforge.WithLogger(logger.Desugar().Sugar()),
	)
	if err != nil {
		return nil, fmt.Errorf("new p2pforge cert manager: %w", err)
	}

	listenAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/webrtc-direct", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/tls/sni/*.%s/ws", cfg.PortWSS, forgeDomain), // cert manager websocket multi address
		fmt.Sprintf("/ip6/::/tcp/%d", cfg.Port),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1", cfg.Port),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1/webtransport", cfg.Port),
		fmt.Sprintf("/ip6/::/udp/%d/webrtc-direct", cfg.Port),
		fmt.Sprintf("/ip6/::/tcp/%d/tls/sni/*.%s/ws", cfg.PortWSS, forgeDomain), // cert manager websocket multi address
	}

	opts := []libp2p.Option{
		libp2p.UserAgent(userAgent),
		libp2p.Identity(cfg.PrivateKey),
		libp2p.Peerstore(ps),
		libp2p.DisableRelay(),
		libp2p.ListenAddrStrings(listenAddrs...),
		libp2p.DisableMetrics(),
		libp2p.ShareTCPListener(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(libp2pwebtransport.New),
		libp2p.Transport(libp2pwebrtc.New),
		libp2p.Transport(libp2pws.New, libp2pws.WithTLSConfig(certMgr.TLSConfig())),
		libp2p.AddrsFactory(certMgr.AddressFactory()),
	}

	if cfg.Port == 0 {
		opts = append(opts, libp2p.NATPortMap()) // enable NAT port mapping if no port is specified
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
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

	if err = dht.Bootstrap(ctx); err != nil {
		logger.Warn("bootstrap failed: %s", err)
	}

	if err = certMgr.Start(); err != nil {
		return nil, fmt.Errorf("start cert manager: %w", err)
	}

	go func() {
		for range certLoadedChan {
			logger.Infow("Loaded certificate", "ant", h.ID())
		}
		logger.Debug("certificate loaded channel closed")
	}()

	sub, err := h.EventBus().Subscribe([]interface{}{new(event.EvtLocalAddressesUpdated), new(event.EvtLocalReachabilityChanged)})
	if err != nil {
		return nil, fmt.Errorf("subscribe to event bus: %w", err)
	}

	go func() {
		for out := range sub.Out() {
			switch evt := out.(type) {
			case event.EvtLocalAddressesUpdated:
				logger.Infow("Addrs updated", "ant", h.ID())
				for i, maddr := range evt.Current {
					actionStr := ""
					switch maddr.Action {
					case event.Added:
						actionStr = "ADD"
					case event.Removed:
						actionStr = "REMOVE"
					default:
						continue
					}
					logger.Infof("  [%d] %s %s", i, actionStr, maddr.Address)
				}
			case event.EvtLocalReachabilityChanged:
				logger.Infow("Reachability changed", "ant", h.ID(), "reachability", evt.Reachability)
			}
		}
	}()

	ant := &Ant{
		cfg:            cfg,
		host:           h,
		dht:            dht,
		certMgr:        certMgr,
		certLoadedChan: certLoadedChan,
		sub:            sub,
		kadID:          PeerIDToKadID(h.ID()),
	}

	return ant, nil
}

func (a *Ant) Close() error {
	if err := a.sub.Close(); err != nil {
		logger.Warnf("failed to close address update subscription: %s", err)
	}

	a.certMgr.Stop()
	close(a.certLoadedChan)

	if err := a.dht.Close(); err != nil {
		logger.Warnf("failed to close dht: %s", err)
	}
	return a.host.Close()
}
