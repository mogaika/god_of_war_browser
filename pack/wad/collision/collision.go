package collision

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mogaika/god_of_war_browser/pack/wad/mat"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const COLLISION_MAGIC = 0x00000011

type Collision struct {
	Magic     uint32
	ShapeName string
	FileSize  uint32
	Shape     interface{}
}

func NewFromData(bs *utils.BufStack, wrtr io.Writer) (c *Collision, err error) {
	head := bs.Raw()[:16]

	c = &Collision{
		Magic: bs.LU32(0),
	}

	for _, sh := range []struct {
		Offset int
		Name   string
	}{{8, "BallHull"}, {4, "SheetHdr"}, {4, "mCDbgHdr"}} {
		if utils.BytesToString(head[sh.Offset:sh.Offset+8]) == sh.Name {
			c.ShapeName = sh.Name
			break
		}
	}

	bs.SetName(c.ShapeName)

	switch c.ShapeName {
	case "SheetHdr":
		var rib *ShapeRibSheet
		if rib, err = NewRibSheet(bs, wrtr); err == nil {
			c.Shape = rib
			if _, err := NewRibSheet(utils.NewBufStack("heh", rib.Marshal()), ioutil.Discard); err != nil {
				log.Printf("Failed to reopen remarshaled file: %v", err)
			}
		}
	case "BallHull":
		c.Shape, err = NewBallHull(bs, wrtr)
	case "mCDbgHdr":
		// Always follow BallHull which it linked to
		c.Shape, err = NewDbgHdr(bs)
	default:
		return nil, fmt.Errorf("Unknown enz shape type %s", c.ShapeName)
	}

	return
}

func (c *Collision) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	if ball, ok := c.Shape.(*ShapeBallHull); ok {
		nextNode := wrsrc.Wad.Nodes[wrsrc.Node.Id+1]
		if strings.ContainsRune(nextNode.Tag.Name, '_') && strings.Split(nextNode.Tag.Name, "_")[1] == strings.Split(wrsrc.Tag.Name, "_")[1] &&
			utils.BytesToString(nextNode.Tag.Data[4:12]) == "mCDbgHdr" {
			var err error
			ball.DbgMesh, err = NewDbgHdr(utils.NewBufStack("mdbgchild", nextNode.Tag.Data))
			if err != nil {
				log.Printf("Failed to parse Collision following mDbg: %v", err)
			}
		}
	} else if rib, ok := c.Shape.(*ShapeRibSheet); ok {
		prefixIndexes := make(map[string]int)
		for iM := range rib.Some4Materials {
			m := &rib.Some4Materials[iM]
			exact := true
			index := 0

			candidats := strings.Split(m.Name, "_")
			suffix := "MAT_" + candidats[len(candidats)-1]
			lastC := rune(suffix[len(suffix)-1])
			if unicode.IsLetter(lastC) && unicode.IsUpper(lastC) {
				exact = false
				suffix = suffix[:len(suffix)-1]
				index = prefixIndexes[suffix]
				prefixIndexes[suffix]++
			}

			var matNode *wad.Node
			if exact {
				matNode = wrsrc.Wad.GetNodeByName(suffix, 0, true)
			} else {
				for _, node := range wrsrc.Wad.Nodes {
					if node.Parent == -1 && strings.HasPrefix(node.Tag.Name, suffix) {
						// log.Println("hit", node.Tag.Name)
						matNode = node
						if index == 0 {
							break
						}
						index -= 1
					}
				}
			}
			if matNode != nil {
				renderM, _, err := wrsrc.Wad.GetInstanceFromNode(matNode.Id)
				if err == nil {
					m.EditorMaterial = matNode.Tag.Name
					colors := renderM.(*mat.Material).Color
					for i := range colors {
						m.EditorColor[i] = colors[i]
					}
				}
				//log.Printf("Found color material %q for physical mat %q", matNode.Tag.Name, m.Name)
			} else {
				log.Printf("Failed to find color material for physical mat %q (exact %t, suffix %q, index %d)",
					m.Name, exact, suffix, index)
			}
		}
	}
	return c, nil
}

func init() {
	wad.SetHandler(config.GOW1, COLLISION_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		fpath := filepath.Join("logs", wrsrc.Wad.Name(), fmt.Sprintf("%.4d-%s.enz.obj", wrsrc.Tag.Id, wrsrc.Tag.Name))
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)
		defer f.Close()

		// f := ioutil.Discard

		bs := utils.NewBufStack("collision", wrsrc.Tag.Data)

		return NewFromData(bs, f)
	})
}
