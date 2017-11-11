package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type FontCharToAsciiByteAssoc map[rune]uint8

func GetFontAliases() (FontCharToAsciiByteAssoc, error) {
	data, err := ioutil.ReadFile("font_aliases.cfg")
	if err != nil {
		return nil, fmt.Errorf("Cannot read file font_aliases.cfg: %v", err)
	}

	var fch map[string]uint8
	if err := json.Unmarshal(data, &fch); err != nil {
		return nil, fmt.Errorf("Unmarshaling error: %v", err)
	}

	rlfch := FontCharToAsciiByteAssoc(make(map[rune]uint8))
	for str, char := range fch {
		rlfch[([]rune(str))[0]] = char
	}
	return rlfch, nil
}
