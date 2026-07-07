package auditctx

import (
	"context"
	"strings"
)

const SystemActor = "system"

type actorKey struct{}

func WithActor(ctx context.Context, actorID string) context.Context {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		actorID = SystemActor
	}
	return context.WithValue(ctx, actorKey{}, actorID)
}

func ActorID(ctx context.Context) string {
	if v, ok := ctx.Value(actorKey{}).(string); ok {
		if id := strings.TrimSpace(v); id != "" {
			return id
		}
	}
	return SystemActor
}