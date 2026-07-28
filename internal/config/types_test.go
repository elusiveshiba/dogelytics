package config

import (
	"encoding/json"
	"testing"
)

func TestDogeAmountUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DogeAmount
		wantErr bool
	}{
		{name: "zero", input: `"0"`, want: "0"},
		{name: "fraction", input: `"12.3456789"`, want: "12.3456789"},
		{name: "larger than int64", input: `"100000000000.00000001"`, want: "100000000000.00000001"},
		{name: "negative", input: `"-1.25"`, want: "-1.25"},
		{name: "number is rejected", input: `1.25`, wantErr: true},
		{name: "exponent is rejected", input: `"1e8"`, wantErr: true},
		{name: "too many decimals", input: `"1.000000001"`, wantErr: true},
		{name: "leading zero", input: `"01"`, wantErr: true},
		{name: "empty", input: `""`, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got DogeAmount
			err := json.Unmarshal([]byte(test.input), &got)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshal amount: %v", err)
			}
			if got != test.want {
				t.Fatalf("amount mismatch: got %q, want %q", got, test.want)
			}
		})
	}
}
