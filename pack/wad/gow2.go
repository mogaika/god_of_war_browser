package wad

import "fmt"

const (
	TAG_GOW2_ENTITY_COUNT     = 0
	TAG_GOW2_SERVER_INSTANCE  = 1
	TAG_GOW2_FILE_GROUP_START = 2
	TAG_GOW2_FILE_GROUP_END   = 3

	TAG_GOW2_HEADER_START = 21
	TAG_GOW2_HEADER_POP   = 19

	TAG_GOW2_TT_11 = 11
	TAG_GOW2_TT_12 = 12
	TAG_GOW2_TT_13 = 13
	TAG_GOW2_TT_14 = 14
	TAG_GOW2_TT_15 = 15
	TAG_GOW2_TT_16 = 16
)

func (w *Wad) gow2parseTag(tag *Tag, currentNode *NodeId, newGroupTag *bool, addNode func(tag *Tag) *Node) error {
	switch tag.Tag {
	case TAG_GOW2_SERVER_INSTANCE: // file data packet
		// Tell file server (server determined by first uint16)
		// that new file is loaded
		// if name start with space, then name ignored (unnamed)
		// overwrite previous instance with same name
		n := addNode(tag)
		if *newGroupTag {
			*newGroupTag = false
			*currentNode = n.Id
		}
	case TAG_GOW2_FILE_GROUP_START:
		*newGroupTag = true
	case TAG_GOW2_FILE_GROUP_END:
		if !*newGroupTag {
			if *currentNode == NODE_INVALID {
				return fmt.Errorf("Trying to end not started group id%d-%s", tag.Id, tag.Name)
			}
			*currentNode = w.GetNodeById(*currentNode).Parent
		} else {
			*newGroupTag = false
		}
	case TAG_GOW2_TT_11, TAG_GOW2_TT_12, TAG_GOW2_TT_13, TAG_GOW2_TT_14, TAG_GOW2_TT_15, TAG_GOW2_TT_16:
		/*
			tag11:
				unk 4 bytes (32 CB 08 4A)
			tag12:
				tweak templates
			tag13:
				map string to tweat template offset? names of tweak templates?
			tag14:
				map string to tewak ttemplate offset? or import/export templates? idk
			tag15:
				map hash to string
			tag16:
				description of some fields in tweak template?
		*/
		addNode(tag)
	case TAG_GOW2_HEADER_START:
		addNode(tag)
	default:
		// return fmt.Errorf("unknown tag id%.4x-tag%.4x-%s offset 0x%.6x", tag.Id, tag.Tag, tag.Name, tag.DebugPos)
	}
	return nil
}
