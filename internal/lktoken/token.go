package lktoken

import (
	"fmt"
	"strings"
	"time"

	"github.com/livekit/protocol/auth"
)

type Config struct {
	APIKey    string
	APISecret string
	LiveURL   string
}

func (c Config) Enabled() bool {
	return strings.TrimSpace(c.APIKey) != "" && strings.TrimSpace(c.APISecret) != ""
}

func (c Config) JoinToken(roomName, identity string, ttl time.Duration) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("livekit is not configured")
	}
	roomName = strings.TrimSpace(roomName)
	identity = strings.TrimSpace(identity)
	if roomName == "" || identity == "" {
		return "", fmt.Errorf("room and identity are required")
	}
	if ttl <= 0 {
		ttl = time.Hour
	}

	at := auth.NewAccessToken(c.APIKey, c.APISecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetValidFor(ttl)

	return at.ToJWT()
}