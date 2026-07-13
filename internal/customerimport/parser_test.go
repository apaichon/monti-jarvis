package customerimport

import (
	"os"
	"strings"
	"testing"
)

func TestParseValidAndInvalidRows(t *testing.T) {
	csv := "display_name,email,tier_slug,group_slugs,source,external_id\nJane,jane@example.com,vip,retail|beta,csv,42\nBad,not-email,,,,\n"
	result, err := Parse(strings.NewReader(csv), 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 || len(result.Rows) != 1 || len(result.Errors) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if got := result.Rows[0].GroupSlugs; len(got) != 2 || got[0] != "retail" {
		t.Fatalf("unexpected groups: %#v", got)
	}
}

func TestParseDryLimitsAndHeader(t *testing.T) {
	if _, err := Parse(strings.NewReader("email\na@example.com\n"), 10); err == nil {
		t.Fatal("expected missing display_name")
	}
	if _, err := Parse(strings.NewReader("display_name,email\nA,a@example.com\nB,b@example.com\n"), 1); err == nil {
		t.Fatal("expected row cap")
	}
}

func TestSampleCustomerImportCSV(t *testing.T) {
	file, err := os.Open("../../docs/samples/customer-import-100.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	result, err := Parse(file, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 100 || len(result.Rows) != 100 || len(result.Errors) != 0 {
		t.Fatalf("sample result total=%d accepted=%d errors=%d, want 100/100/0", result.Total, len(result.Rows), len(result.Errors))
	}
}
