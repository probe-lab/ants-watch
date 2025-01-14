package db

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Request struct {
	UUID               uuid.UUID              `ch:"id"`
	QueenID            string                 `ch:"queen_id"`
	AntID              peer.ID                `ch:"ant_multihash"`
	RemoteID           peer.ID                `ch:"remote_multihash"`
	RequestType        pb.Message_MessageType `ch:"request_type"`
	AgentVersion       string                 `ch:"agent_version"`
	AgentVersionType   string                 `ch:"agent_version_type"`
	AgentVersionSemVer [3]int                 `ch:"agent_version_semver"`
	Protocols          []string               `ch:"protocols"`
	StartedAt          time.Time              `ch:"started_at"`
	KeyID              string                 `ch:"key_multihash"`
	MultiAddresses     []string               `ch:"multi_addresses"`
	ConnMaddr          string                 `ch:"conn_maddr"`
}
