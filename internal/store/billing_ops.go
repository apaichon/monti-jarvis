package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/libra/monti-jarvis/internal/auditctx"
)

// ListPaymentOrders returns a platform billing ledger (Sprint 10).
func (s *Store) ListPaymentOrders(ctx context.Context, f PaymentOrderListFilter) ([]PaymentOrderListItem, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if tid := strings.TrimSpace(f.TenantID); tid != "" {
		where = append(where, fmt.Sprintf("o.tenant_id = $%d", n))
		args = append(args, tid)
		n++
	}
	if st := strings.TrimSpace(f.Status); st != "" {
		where = append(where, fmt.Sprintf("o.status = $%d", n))
		args = append(args, st)
		n++
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`
SELECT o.id, o.tenant_id, o.package_id, o.order_no, o.amount_cents, o.currency, o.status, o.provider,
       COALESCE(o.payment_method, 'credit_card'), o.transaction_id, o.payment_url, o.paid_at, o.created_at, o.updated_at,
       COALESCE(p.name, o.package_id), COALESCE(t.name, o.tenant_id)
FROM %s.payment_orders o
LEFT JOIN %s.packages p ON p.id = o.package_id
LEFT JOIN %s.tenants t ON t.id = o.tenant_id
WHERE %s
ORDER BY o.created_at DESC
LIMIT $%d OFFSET $%d`, schema, schema, schema, strings.Join(where, " AND "), n, n+1)

	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PaymentOrderListItem
	for rows.Next() {
		var item PaymentOrderListItem
		if err := rows.Scan(
			&item.ID, &item.TenantID, &item.PackageID, &item.OrderNo, &item.AmountCents, &item.Currency,
			&item.Status, &item.Provider, &item.PaymentMethod, &item.TransactionID, &item.PaymentURL,
			&item.PaidAt, &item.CreatedAt, &item.UpdatedAt, &item.PackageName, &item.TenantName,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListPaymentDocuments filters platform documents (Sprint 11).
func (s *Store) ListPaymentDocuments(ctx context.Context, f PaymentDocumentListFilter) ([]PaymentDocument, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if tid := strings.TrimSpace(f.TenantID); tid != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", n))
		args = append(args, tid)
		n++
	}
	if dt := strings.TrimSpace(f.DocType); dt != "" {
		where = append(where, fmt.Sprintf("doc_type = $%d", n))
		args = append(args, dt)
		n++
	}
	if st := strings.TrimSpace(f.Status); st != "" {
		where = append(where, fmt.Sprintf("COALESCE(status, 'issued') = $%d", n))
		args = append(args, st)
		n++
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`
SELECT %s FROM %s.payment_documents
WHERE %s
ORDER BY issued_at DESC
LIMIT $%d OFFSET $%d`, paymentDocumentSelectCols, schema, strings.Join(where, " AND "), n, n+1)

	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PaymentDocument
	for rows.Next() {
		d, err := scanPaymentDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListTenantPaymentDocuments lists documents for one tenant (Sprint 12 vault).
func (s *Store) ListTenantPaymentDocuments(ctx context.Context, tenantID string) ([]PaymentDocument, error) {
	return s.ListPaymentDocuments(ctx, PaymentDocumentListFilter{TenantID: tenantID, Limit: 100})
}

// VoidPaymentDocument marks a document voided (Sprint 11).
func (s *Store) VoidPaymentDocument(ctx context.Context, docID, reason string) (*PaymentDocument, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	doc, err := s.GetPaymentDocumentByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.Status == PaymentDocStatusVoided {
		return doc, nil
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_documents
SET status = 'voided', void_reason = $2, voided_at = now(), updated_by = $3
WHERE id = $1 AND COALESCE(status, 'issued') = 'issued'`, schema),
		docID, strings.TrimSpace(reason), actor)
	if err != nil {
		return nil, err
	}
	return s.GetPaymentDocumentByID(ctx, docID)
}

// ReissuePaymentDocument voids the current active doc of the same type and issues a new one (Sprint 11).
func (s *Store) ReissuePaymentDocument(ctx context.Context, docID, reason string) (*PaymentDocument, error) {
	if s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	old, err := s.GetPaymentDocumentByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if old.Status == PaymentDocStatusVoided {
		// Find active sibling or re-issue from order.
		active, aerr := s.GetPaymentDocument(ctx, old.OrderID, old.DocType)
		if aerr == nil {
			return s.ReissuePaymentDocument(ctx, active.ID, reason)
		}
	} else {
		if _, err := s.VoidPaymentDocument(ctx, docID, reason); err != nil {
			return nil, err
		}
	}

	order, err := s.GetPaymentOrderByID(ctx, old.OrderID)
	if err != nil {
		return nil, err
	}
	if order.Status != PaymentOrderStatusPaid {
		return nil, fmt.Errorf("order is not paid")
	}
	pkg, err := s.GetPackage(ctx, order.PackageID)
	if err != nil {
		return nil, err
	}
	buyerName, buyerAddr, buyerTaxID := s.resolveBuyerFields(ctx, order.TenantID)
	seller, _ := s.GetSellerBranding(ctx)
	net, vat := splitVATInclusive(order.AmountCents, old.VATRateBps)
	if old.VATRateBps <= 0 {
		net, vat = splitVATInclusive(order.AmountCents, 700)
	}
	vatRate := old.VATRateBps
	if vatRate <= 0 {
		vatRate = 700
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	prefix := "RCP"
	if old.DocType == PaymentDocTypeTaxInvoice {
		prefix = "TAX"
	}
	id := "pdoc_" + newStoreID()
	docNo := fmt.Sprintf("%s-%s-%s", prefix, order.OrderNo, newStoreID()[:8])
	issuedAt := time.Now().UTC()
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_documents (
  id, order_id, tenant_id, doc_type, doc_number, status,
  buyer_name, buyer_address, buyer_tax_id,
  seller_name, seller_address, seller_tax_id,
  package_name, amount_cents, currency, vat_rate_bps, net_cents, vat_cents,
  payment_method, reissued_from_id, issued_at, created_by, updated_by
) VALUES (
  $1,$2,$3,$4,$5,'issued',
  $6,$7,$8,
  $9,$10,$11,
  $12,$13,$14,$15,$16,$17,
  $18,$19,$20,$21,$21
)`, schema),
		id, order.ID, order.TenantID, old.DocType, docNo,
		buyerName, buyerAddr, buyerTaxID,
		seller.Name, seller.Address, seller.TaxID,
		pkg.Name, order.AmountCents, order.Currency, vatRate, net, vat,
		order.PaymentMethod, old.ID, issuedAt, actor,
	)
	if err != nil {
		return nil, err
	}
	return s.GetPaymentDocumentByID(ctx, id)
}

// GetSellerBranding returns platform seller block (Sprint 11).
func (s *Store) GetSellerBranding(ctx context.Context) (PlatformSellerBranding, error) {
	if s.pg == nil {
		return PlatformSellerBranding{
			ID: PlatformSellerBrandingID, Name: "Monti Jarvis Platform",
			Address: "Bangkok, Thailand", TaxID: "0-0000-00000-00-0", Branch: "00000",
		}, nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var b PlatformSellerBranding
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, name, address, tax_id, branch, updated_at
FROM %s.platform_seller_branding WHERE id = $1`, schema), PlatformSellerBrandingID).
		Scan(&b.ID, &b.Name, &b.Address, &b.TaxID, &b.Branch, &b.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PlatformSellerBranding{
			ID: PlatformSellerBrandingID, Name: "Monti Jarvis Platform",
			Address: "Bangkok, Thailand", TaxID: "0-0000-00000-00-0", Branch: "00000",
		}, nil
	}
	if err != nil {
		return PlatformSellerBranding{}, err
	}
	return b, nil
}

// UpsertSellerBranding updates platform seller block.
func (s *Store) UpsertSellerBranding(ctx context.Context, name, address, taxID, branch string) (PlatformSellerBranding, error) {
	if s.pg == nil {
		return PlatformSellerBranding{}, errors.New("postgres unavailable")
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	if strings.TrimSpace(name) == "" {
		name = "Monti Jarvis Platform"
	}
	if strings.TrimSpace(branch) == "" {
		branch = "00000"
	}
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.platform_seller_branding (id, name, address, tax_id, branch, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$6)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  address = EXCLUDED.address,
  tax_id = EXCLUDED.tax_id,
  branch = EXCLUDED.branch,
  updated_by = EXCLUDED.updated_by,
  updated_at = now()`, schema),
		PlatformSellerBrandingID, strings.TrimSpace(name), strings.TrimSpace(address),
		strings.TrimSpace(taxID), strings.TrimSpace(branch), actor)
	if err != nil {
		return PlatformSellerBranding{}, err
	}
	return s.GetSellerBranding(ctx)
}

// GetTenantTaxProfile returns buyer tax fields (Sprint 12).
func (s *Store) GetTenantTaxProfile(ctx context.Context, tenantID string) (TenantTaxProfile, error) {
	if s.pg == nil {
		return TenantTaxProfile{TenantID: tenantID}, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var p TenantTaxProfile
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT tenant_id, company_name, tax_id, branch, address, updated_at
FROM %s.tenant_tax_profiles WHERE tenant_id = $1`, schema), tenantID).
		Scan(&p.TenantID, &p.CompanyName, &p.TaxID, &p.Branch, &p.Address, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Fall back to registration/KYC defaults without error.
		p = TenantTaxProfile{TenantID: tenantID, Branch: "00000"}
		if reg, regErr := s.GetTenantRegistration(ctx, tenantID); regErr == nil {
			p.CompanyName = reg.CompanyName
		}
		if kyc, kycErr := s.GetTenantKYCProfile(ctx, tenantID); kycErr == nil {
			p.Address = kyc.ContactAddress
		}
		return p, nil
	}
	if err != nil {
		return TenantTaxProfile{}, err
	}
	return p, nil
}

// UpsertTenantTaxProfile saves buyer tax fields.
func (s *Store) UpsertTenantTaxProfile(ctx context.Context, tenantID string, companyName, taxID, branch, address string) (TenantTaxProfile, error) {
	if s.pg == nil {
		return TenantTaxProfile{}, errors.New("postgres unavailable")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return TenantTaxProfile{}, errors.New("tenant_id required")
	}
	if strings.TrimSpace(branch) == "" {
		branch = "00000"
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_tax_profiles (tenant_id, company_name, tax_id, branch, address, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$6)
ON CONFLICT (tenant_id) DO UPDATE SET
  company_name = EXCLUDED.company_name,
  tax_id = EXCLUDED.tax_id,
  branch = EXCLUDED.branch,
  address = EXCLUDED.address,
  updated_by = EXCLUDED.updated_by,
  updated_at = now()`, schema),
		tenantID, strings.TrimSpace(companyName), strings.TrimSpace(taxID),
		strings.TrimSpace(branch), strings.TrimSpace(address), actor)
	if err != nil {
		return TenantTaxProfile{}, err
	}
	return s.GetTenantTaxProfile(ctx, tenantID)
}

// RefreshTenantTaxInvoices reissues active tax invoices for the tenant with updated tax profile.
func (s *Store) RefreshTenantTaxInvoices(ctx context.Context, tenantID string) (int, error) {
	docs, err := s.ListPaymentDocuments(ctx, PaymentDocumentListFilter{
		TenantID: tenantID,
		DocType:  PaymentDocTypeTaxInvoice,
		Status:   PaymentDocStatusIssued,
		Limit:    100,
	})
	if err != nil {
		return 0, err
	}
	n := 0
	for _, d := range docs {
		if _, err := s.ReissuePaymentDocument(ctx, d.ID, "tax profile updated"); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
