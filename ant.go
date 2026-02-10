package ants

import (
	"context"
	"fmt"
	"time"

	"github.com/caddyserver/certmagic"
	ds "github.com/ipfs/go-datastore"
	p2pforge "github.com/ipshipyard/p2p-forge/client"
	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	libp2ptcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pwebrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	libp2pws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	libp2pwebtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"github.com/ipfs/go-libdht/kad/key/bit256"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/probe-lab/ants-watch/metrics"
)

type RequestEvent struct {
	Timestamp    time.Time
	Self         peer.ID
	Remote       peer.ID
	Type         pb.Message_MessageType
	Target       mh.Multihash
	AgentVersion string
	Protocols    []protocol.ID
	Maddrs       []multiaddr.Multiaddr
	ConnMaddr    multiaddr.Multiaddr
}

func (r *RequestEvent) IsIdentified() bool {
	return r.AgentVersion != "" && len(r.Protocols) > 0
}

func (r *RequestEvent) MaddrStrings() []string {
	maddrStrs := make([]string, len(r.Maddrs))
	for i, maddr := range r.Maddrs {
		maddrStrs[i] = maddr.String()
	}
	return maddrStrs
}

type AntConfig struct {
	PrivateKey     crypto.PrivKey
	UserAgent      string
	Port           int
	ProtocolID     string
	BootstrapPeers []peer.AddrInfo
	RequestsChan   chan<- RequestEvent
	CertPath       string
	Telemetry      *metrics.Telemetry
}

func (cfg *AntConfig) Validate() error {
	if cfg.PrivateKey == nil {
		return fmt.Errorf("no ant private key given")
	}

	if cfg.UserAgent == "" {
		return fmt.Errorf("user agent is not set")
	}

	if cfg.ProtocolID == "" {
		return fmt.Errorf("protocol prefix is not set")
	}

	if len(cfg.BootstrapPeers) == 0 {
		return fmt.Errorf("bootstrap peers are not set")
	}

	if cfg.RequestsChan == nil {
		return fmt.Errorf("requests channel is not set")
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
		p2pforge.WithLogger(logger.Desugar().WithOptions(zap.IncreaseLevel(zap.InfoLevel)).Sugar()),
		p2pforge.WithCertificateStorage(&certmagic.FileStorage{Path: cfg.CertPath}),
	)
	if err != nil {
		return nil, fmt.Errorf("new p2pforge cert manager: %w", err)
	}

	// Configure the resource manager to not limit anything
	noSubnetLimit := []rcmgr.ConnLimitPerSubnet{}
	noNetPrefixLimit := []rcmgr.NetworkPrefixLimit{}
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter,
		rcmgr.WithLimitPerSubnet(noSubnetLimit, noSubnetLimit),
		rcmgr.WithNetworkPrefixLimit(noNetPrefixLimit, noNetPrefixLimit),
	)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	listenAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/webrtc-direct", cfg.Port),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/tls/sni/*.%s/ws", cfg.Port, forgeDomain), // cert manager websocket multi address
		fmt.Sprintf("/ip6/::/tcp/%d", cfg.Port),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1", cfg.Port),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1/webtransport", cfg.Port),
		fmt.Sprintf("/ip6/::/udp/%d/webrtc-direct", cfg.Port),
		fmt.Sprintf("/ip6/::/tcp/%d/tls/sni/*.%s/ws", cfg.Port, forgeDomain), // cert manager websocket multi address
	}

	opts := []libp2p.Option{
		libp2p.UserAgent(cfg.UserAgent),
		libp2p.Identity(cfg.PrivateKey),
		libp2p.Peerstore(ps),
		libp2p.DisableRelay(),
		libp2p.ListenAddrStrings(listenAddrs...),
		libp2p.DisableMetrics(),
		libp2p.ShareTCPListener(),
		libp2p.ResourceManager(rm),
		libp2p.ConnectionManager(connmgr.NullConnMgr{}),
		libp2p.Transport(libp2ptcp.NewTCPTransport),
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

	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			cfg.Telemetry.ConnectCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("direction", conn.Stat().Direction.String()),
			))
		},
		DisconnectedF: func(n network.Network, conn network.Conn) {
			cfg.Telemetry.DisconnectCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("direction", conn.Stat().Direction.String()),
			))
		},
	})

	dhtOpts := []kad.Option{
		kad.Mode(kad.ModeServer),
		kad.BootstrapPeers(cfg.BootstrapPeers...),
		kad.V1ProtocolOverride(protocol.ID(cfg.ProtocolID)),
		kad.Datastore(ds),
		kad.OnRequestHook(onRequestHook(h, cfg)),
	}
	dht, err := kad.New(ctx, h, dhtOpts...)
	if err != nil {
		return nil, fmt.Errorf("new libp2p dht: %w", err)
	}
	logger.Debugf("spawned ant. kadid: %s, peerid: %s", PeerIDToKadID(h.ID()).HexString(), h.ID())

	if err = dht.Bootstrap(ctx); err != nil {
		logger.Warn("bootstrap failed: %s", err)
	}

	certMgr.ProvideHost(h)

	if err = certMgr.Start(); err != nil {
		return nil, fmt.Errorf("start cert manager: %w", err)
	}

	go func() {
		for range certLoadedChan {
			logger.Infow("Loaded certificate", "ant", h.ID())
		}
		logger.Debug("certificate loaded channel closed")
	}()

	sub, err := h.EventBus().Subscribe([]interface{}{
		new(event.EvtLocalAddressesUpdated),
		new(event.EvtLocalReachabilityChanged),
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe to event bus: %w", err)
	}

	go func() {
		for out := range sub.Out() {
			switch evt := out.(type) {
			case event.EvtLocalAddressesUpdated:
				if !evt.Diffs {
					continue
				}

				logger.Infow("Ant now listening on:", "ant", h.ID())
				for i, maddr := range evt.Current {
					actionStr := ""
					switch maddr.Action {
					case event.Added:
						actionStr = "ADD"
					case event.Removed:
						actionStr = "REMOVE"
					case event.Maintained:
						actionStr = "MAINTAINED"
					default:
						continue
					}
					logger.Infof("[%d] %s %s/p2p/%s", i, actionStr, maddr.Address, h.ID())
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

func onRequestHook(h host.Host, cfg *AntConfig) func(ctx context.Context, s network.Stream, req pb.Message) {
	return func(ctx context.Context, s network.Stream, req pb.Message) {
		remotePeer := s.Conn().RemotePeer()

		agentVersion := ""
		val, err := h.Peerstore().Get(remotePeer, "AgentVersion")
		if err == nil {
			agentVersion = val.(string)
		}

		maddrs := h.Peerstore().Addrs(remotePeer)
		protocolIDs, _ := h.Peerstore().GetProtocols(remotePeer) // ignore error

		cfg.RequestsChan <- RequestEvent{
			Timestamp:    time.Now(),
			Self:         h.ID(),
			Remote:       remotePeer,
			Type:         req.GetType(),
			Target:       req.GetKey(),
			AgentVersion: agentVersion,
			Protocols:    protocolIDs,
			Maddrs:       maddrs,
			ConnMaddr:    s.Conn().RemoteMultiaddr(),
		}
	}
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
