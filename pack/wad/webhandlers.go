package wad

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

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

func (wad *Wad) WebHandlerDumpTagData(w http.ResponseWriter, id TagId) {
	tag := wad.GetTagById(id)
	node := wad.GetNodeById(tag.NodeId)
	webutils.WriteFile(w, bytes.NewBuffer(node.Tag.Data), node.Tag.Name)
}

func (wad *Wad) WebHandlerCallResourceHttpAction(w http.ResponseWriter, r *http.Request, id TagId, action string) error {
	switch action {
	case "updatetag":
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("Cannot parse form: %v", err)
		}

		tagTag, err := strconv.ParseInt(r.Form.Get("tagtag"), 0, 16)
		if err != nil {
			return err
		}
		tagName := r.Form.Get("tagname")
		tagFlags, err := strconv.ParseInt(r.Form.Get("tagflags"), 0, 16)
		if err != nil {
			return err
		}

		if err := wad.UpdateTagInfo(map[TagId]Tag{id: {Id: id, Tag: uint16(tagTag), Flags: uint16(tagFlags), Name: tagName}}); err != nil {
			return fmt.Errorf("Error when updating wad tag %d: %v", id, err)
		}
	default:
		if inst, _, err := wad.GetInstanceFromTag(id); err == nil {
			rt := reflect.TypeOf(inst)
			method, has := rt.MethodByName("HttpAction")
			if !has {
				return fmt.Errorf("Error: %s has not func SubfileGetter", rt.Name())
			} else {
				method.Func.Call([]reflect.Value{
					reflect.ValueOf(inst),
					reflect.ValueOf(wad.GetNodeResourceByTagId(id)),
					reflect.ValueOf(w),
					reflect.ValueOf(r),
					reflect.ValueOf(action),
				}[:])
				return nil
			}
		} else {
			return fmt.Errorf("File %d instance getting error: %v", id, err)
		}
	}
	return nil
}
