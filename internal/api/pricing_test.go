package api

import "testing"

func TestParsePrice(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"4.35", 4.35},
		{"0.0440", 0.044},
		{"", 0},
		{"nope", 0},
	}
	for _, tc := range cases {
		if got := parsePrice(tc.in); got != tc.want {
			t.Fatalf("parsePrice(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestLookupServerPrice(t *testing.T) {
	pricing := SampleInventory().Pricing
	if got := lookupServerPrice(pricing, "CX22", "fsn1"); got != 4.35 {
		t.Fatalf("type-only fallback = %v, want 4.35", got)
	}
	pricing.ServerMonthly["CX22|fsn1"] = 4.99
	if got := lookupServerPrice(pricing, "CX22", "fsn1"); got != 4.99 {
		t.Fatalf("location price = %v, want 4.99", got)
	}
}
