package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type tenantExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func setTenantContext(ctx context.Context, exec tenantExecutor, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return fmt.Errorf("tenant context is required")
	}
	_, err := exec.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantID)
	return err
}
