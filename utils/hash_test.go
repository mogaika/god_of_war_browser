package utils

import "testing"

var hashTests = []struct {
	in_str    string
	in_init   uint32
	out_nodes uint32
}{
	{"", 0, 0x0},
	{"@", 0, 0x40},
	{"010", 0, 0xbe8af},
	{"014", 0, 0xbe8b3},
	{"System", 0, 0xd94d08af},
	{"Debug", 0, 0x2ad2a833},
	{"WadInfoGroup", 0, 0x228c2437},
	{"TweakTemplates", 0, 0x372a9aad},
}

func TestStringHashNodes(t *testing.T) {
	for _, test := range hashTests {
		result := GameStringHashNodes(test.in_str, test.in_init)
		if result != test.out_nodes {
			t.Errorf("GameStringHashNodes(%q,%d)=%d; expected %d", test.in_str, test.in_init, result, test.out_nodes)
		}
	}
}
