package sbk

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/web/webutils"
)

const SBK_SBLK_MAGIC = 0x18
const SBK_VAG_MAGIC = 0x40018

type Sound struct {
	Name     string
	StreamId uint32 // file offset for vag,
}

type SBK struct {
	Sounds          []Sound
	IsVagFiles      bool
	IsSblkContainer bool
}

func NewFromData(f io.ReaderAt, isSblk bool) (*SBK, error) {
	var soundsCount uint32
	if err := binary.Read(io.NewSectionReader(f, 4, 8), binary.LittleEndian, &soundsCount); err != nil {
		return nil, err
	}

	log.Println("Sounds loaded: ", soundsCount)

	sbk := &SBK{
		Sounds:          make([]Sound, soundsCount),
		IsSblkContainer: isSblk,
		IsVagFiles:      !isSblk,
	}

	for i := range sbk.Sounds {
		var buf [28]byte
		_, err := f.ReadAt(buf[:], 8+int64(i)*28)
		if err != nil {
			return nil, err
		}

		sbk.Sounds[i].Name = utils.BytesToString(buf[:24])
		sbk.Sounds[i].StreamId = binary.LittleEndian.Uint32(buf[24:])
	}

	return sbk, nil
}

func (sbk *SBK) SubfileGetter(w http.ResponseWriter, r *http.Request, wad *wad.Wad, node *wad.WadNode, subfile string) {
	rdr, err := wad.GetFileReader(node.Id)
	if err != nil {
		panic(err)
	}

	needWav := false
	if strings.HasSuffix(subfile, "@wav@") {
		subfile = strings.TrimSuffix(subfile, "@wav@")
		needWav = true
	}
	log.Println(subfile)

	if sbk.IsVagFiles {
		for iSnd, snd := range sbk.Sounds {
			if snd.Name == subfile {
				end := node.Size
				if iSnd != len(sbk.Sounds)-1 {
					end = sbk.Sounds[iSnd+1].StreamId
				}

				vagpReader := io.NewSectionReader(rdr, int64(snd.StreamId), int64(end-snd.StreamId))
				if needWav {
					vag, err := vagp.NewVAGPFromReader(vagpReader)
					if err != nil {
						webutils.WriteError(w, err)
					} else {
						wav, err := vag.AsWave()
						if err != nil {
							webutils.WriteError(w, err)
						} else {
							webutils.WriteFile(w, wav, subfile+".WAV")
						}
					}
				} else {
					webutils.WriteFile(w, vagpReader, subfile+".VAG")
				}
				return
			}
		}
		webutils.WriteError(w, errors.New("Cannot find sound"))
	} else {
		start := int64(8 + len(sbk.Sounds)*28)
		webutils.WriteFile(w, io.NewSectionReader(rdr, start, int64(node.Size)-start), node.Name+".SBK")
	}
}

func (sbk *SBK) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return sbk, nil
}

func init() {
	wad.SetHandler(SBK_SBLK_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r, true)
	})
	wad.SetHandler(SBK_VAG_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r, false)
	})
}
