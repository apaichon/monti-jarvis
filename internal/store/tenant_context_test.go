package store

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type captureTenantExec struct {
	query string
	args  []any
}

func (f *captureTenantExec) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	f.query = query
	f.args = args
	return pgconn.NewCommandTag("SELECT 1"), nil
}

func TestSetTenantContextUsesTransactionLocalSetting(t *testing.T) {
	exec := &captureTenantExec{}
	if err := setTenantContext(context.Background(), exec, "tenant-a"); err != nil {
		t.Fatalf("set tenant context: %v", err)
	}
	if exec.query != `SELECT set_config('app.tenant_id', $1, true)` {
		t.Fatalf("query = %q", exec.query)
	}
	if len(exec.args) != 1 || exec.args[0] != "tenant-a" {
		t.Fatalf("args = %#v", exec.args)
	}
}

func TestSetTenantContextRejectsMissingTenant(t *testing.T) {
	if err := setTenantContext(context.Background(), &captureTenantExec{}, " "); err == nil {
		t.Fatal("expected missing tenant context error")
	}
}
