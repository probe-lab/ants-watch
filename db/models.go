package db

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Request struct {
	UUID           uuid.UUID
	QueenID        string
	AntID          peer.ID
	RemoteID       peer.ID
	Type           pb.Message_MessageType
	AgentVersion   string
	Protocols      []string
	StartedAt      time.Time
	KeyID          string
	MultiAddresses []string
}
