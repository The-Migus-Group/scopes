package dhcp

import (
	"strings"
	"testing"
)

func TestOptionNames(t *testing.T) {
	if val, ok := optionNames[3]; !ok {
		t.Error("Expected option 3 to be defined but it was not.")
	} else if val != "Router" {
		t.Errorf("Expected option 3 to be 'Router' but it was '%s'.", val)
	}
}

func TestLoadOptionNamesCsv(t *testing.T) {
	reader := strings.NewReader(`
Tag,Name,"Data
            Length",Meaning,Reference
0,Pad,0,None,[RFC2132]
1,Subnet Mask,4,Subnet Mask Value,[RFC2132]
15,Domain Name,N,The DNS domain name of the client,[RFC2132]`)
	if err := LoadOptionNameCsv(reader); err != nil {
		t.Errorf("Expected no error but got '%s'.", err)
	}
	if val, ok := optionNames[16]; ok {
		t.Errorf("Expected option 16 to be undefined but it was '%s'.", val)
	}
	if val, ok := optionNames[15]; !ok {
		t.Error("Expected option 15 to be defined but it was not.")
	} else if val != "Domain Name" {
		t.Errorf("Expected option 15 to be 'Domain Name' but it was '%s'.", val)
	}
}
