package microsoft

import "testing"

func TestDHCPOptionType(t *testing.T) {
	tests := []struct {
		name       string
		optionType uint8
		want       string
	}{
		{"DWORD", DHCPOptionTypeDWORD, "DWord"},
		{"STRING", DHCPOptionTypeSTRING, "String"},
		{"IPADDRESS", DHCPOptionTypeIPADDRESS, "IPAddress"},
		{"Unknown", 255, "Unknown (255)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DHCPOptionType(tt.optionType); got != tt.want {
				t.Errorf("OptionType() = %v, want %v", got, tt.want)
			}
		})
	}
}
