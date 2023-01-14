package twk

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/twk/twktree"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (t *TWK) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "asyaml":
		fake := *t

		if tree, err := twktree.Root().UnmarshalTWK(t.Tree); err != nil {
			log.Printf("Failed to produce abstract tree: %v", err)
		} else {
			fake.AbstractTree = tree
			fake.Tree = nil
			log.Printf("Produced abstract tree")
		}

		var buffer bytes.Buffer
		enc := yaml.NewEncoder(&buffer)
		enc.SetIndent(2)

		if err := enc.Encode(&fake); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to marshal yaml"))
			return
		}

		if err := enc.Close(); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to close yaml encoder"))
			return
		}

		webutils.WriteFile(w, &buffer, fmt.Sprintf("%s-%d-%s.yaml", wrsrc.Wad.Name(), wrsrc.Tag.Id, wrsrc.Name()))
		return
	case "fromyaml":
		if strings.ToUpper(r.Method) != "POST" {
			return
		}

		eff, _, err := r.FormFile("data")
		if err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to open file"))
			return
		}
		defer eff.Close()

		var fake TWK
		if err := yaml.NewDecoder(eff).Decode(&fake); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to unmarshal yaml"))
			return
		}

		if fake.AbstractTree != nil {
			// convert abstact tree json repr to tree
			root, err := twktree.Root().MarshalTWK(fake.AbstractTree)
			if err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to convert abstract tree"))
				return
			}
			fake.Tree = root
		}

		var buffer bytes.Buffer
		if err := fake.Produce(&buffer); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to produce binary"))
			return
		}

		utils.LogDump(buffer.Bytes())

		if _, err := NewTwkFromData(utils.NewBufStack("testtwk", buffer.Bytes())); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed sanity check"))
			return
		}

		if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
			wrsrc.Tag.Id: buffer.Bytes(),
		}); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to write tag"))
			return
		}
	}
}
