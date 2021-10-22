package config

import (
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
)

var cuurentCharMap *charmap.Charmap = charmap.Windows1252

func SetEncoding(name string) error {
	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			if cm.String() == name {
				cuurentCharMap = cm
				return nil
			}
		}
	}
	return errors.Errorf("Failed to find encoding %q", name)
}

func ListEncodings() []string {
	list := make([]string, 0)
	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			list = append(list, cm.String())
		}
	}
	return list
}

func GetEncoding() *charmap.Charmap {
	return cuurentCharMap
}
