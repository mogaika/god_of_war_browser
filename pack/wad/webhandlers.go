package wad

import (
	"fmt"
	"net/http"

	"github.com/mogaika/god_of_war_browser/webutils"
)

func (wad *Wad) WebHandlerForNodeByTagId(w http.ResponseWriter, tagId TagId) error {
	tag := wad.GetTagById(tagId)
	node := wad.GetNodeById(tag.NodeId)
	data, serverId, err := wad.GetInstanceFromNode(node.Id)
	if err == nil {
		type Result struct {
			Tag      *Tag
			Data     interface{}
			ServerId uint32
		}
		val, err := data.Marshal(wad.GetNodeResourceByTagId(node.Tag.Id))
		if err != nil {
			return fmt.Errorf("Error marshaling node %d from %s: %v", tagId, wad.Name(), err.(error))
		} else {
			webutils.WriteJson(w, &Result{Tag: node.Tag, Data: val, ServerId: serverId})
		}
	} else {
		return fmt.Errorf("File %s-%d[%s] parsing error: %v", wad.Name(), node.Tag.Id, node.Tag.Name, err)
	}
	return nil
}
