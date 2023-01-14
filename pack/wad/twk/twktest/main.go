package main

import (
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad/twk"
	"github.com/mogaika/god_of_war_browser/pack/wad/twk/twktree"
	"github.com/mogaika/god_of_war_browser/utils"
)

func main() {
	//r := new(twktree.Root)
	r := new(twktree.UnknownTreeElement)

	template := &twk.TWK{
		MagicHeaderPresened: true, HeaderStrangeMagicUid: 0x1,
		Directory: twk.Directory{
			Name: "/TweakTemplates/ForceFeedback/123/", Values: []twk.Value{
				twk.Value{Name: "Instance Name", Hex: "736d616c6c206869740000000000000000000000000000000000000000000000", Offset: 49},
				twk.Value{Name: "Duration", Hex: "cdcccc3d", Offset: 86}, twk.Value{Name: "SM frequency", Hex: "00007042", Offset: 95},
				twk.Value{Name: "SM amplitude", Hex: "0000003f", Offset: 104}, twk.Value{Name: "SM phase", Hex: "00000000", Offset: 113},
				twk.Value{Name: "SM bias", Hex: "00000000", Offset: 122}, twk.Value{Name: "SM Waveform", Hex: "01000000", Offset: 131},
				twk.Value{Name: "LM frequency", Hex: "00007042", Offset: 140}, twk.Value{Name: "LM amplitude", Hex: "00007f43", Offset: 149},
				twk.Value{Name: "LM phase", Hex: "0000803f", Offset: 158}, twk.Value{Name: "LM bias", Hex: "0000803f", Offset: 167},
				twk.Value{Name: "LM Waveform", Hex: "0200803f", Offset: 176}, twk.Value{Name: "Attack", Hex: "8fc2f53c", Offset: 185},
				twk.Value{Name: "Decay", Hex: "0ad7233c", Offset: 194}, twk.Value{Name: "Sustain", Hex: "00000000", Offset: 203},
				twk.Value{Name: "Release", Hex: "00000000", Offset: 212}, twk.Value{Name: "Input 0 Bias", Hex: "00000000", Offset: 221},
				twk.Value{Name: "Input 0 Scale", Hex: "00000040", Offset: 230}, twk.Value{Name: "Input 0 Operator", Hex: "01000040", Offset: 239},
				twk.Value{Name: "Input 1 Bias", Hex: "00000000", Offset: 248}, twk.Value{Name: "Input 1 Scale", Hex: "00000000", Offset: 257},
				twk.Value{Name: "Input 1 Operator", Hex: "00000000", Offset: 266}, twk.Value{Name: "Input 2 Bias", Hex: "00000000", Offset: 275},
				twk.Value{Name: "Input 2 Scale", Hex: "00002041", Offset: 284}, twk.Value{Name: "Input 2 Operator", Hex: "00002041", Offset: 293},
				twk.Value{Name: "Input 3 Bias", Hex: "00000000", Offset: 302}, twk.Value{Name: "Input 3 Scale", Hex: "00000000", Offset: 311},
				twk.Value{Name: "Input 3 Operator", Hex: "00000000", Offset: 320},
			},
			Directories: []*twk.Directory(nil),
		},
	}

	defer utils.LogDump(r)
	if err := twktree.AddTWK(r, template); err != nil {
		log.Printf("Error: %v", err)
	}

	/*if data, err := json.MarshalIndent(r, "", "  "); err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Println(string(data))
	}*/
}
