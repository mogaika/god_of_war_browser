package wad

import "fmt"

const (
	TAG_GOW2018_SERVER_INSTANCE  = 1
	TAG_GOW2018_FILE_GROUP_START = 2
	TAG_GOW2018_FILE_GROUP_END   = 3

	TAG_GOW2018_DCClientGUID = 7
	TAG_GOW2018_AUTOPAD      = 0x19 // pad to 0x10000 bytes

	TAG_GOW2018_SIZE = 0x60
)

func (w *Wad) gow2018parseTag(tag *Tag, currentNode *NodeId, newGroupTag *bool, addNode func(tag *Tag) *Node) error {
	switch tag.Tag {
	case TAG_GOW2018_SERVER_INSTANCE:
		n := addNode(tag)
		if *newGroupTag {
			*newGroupTag = false
			*currentNode = n.Id
		}
	case TAG_GOW2018_FILE_GROUP_START:
		*newGroupTag = true
	case TAG_GOW2018_FILE_GROUP_END:
		if !*newGroupTag {
			if *currentNode == NODE_INVALID {
				return fmt.Errorf("Trying to end not started group id%d-%s", tag.Id, tag.Name)
			}
			*currentNode = w.GetNodeById(*currentNode).Parent
		} else {
			*newGroupTag = false
		}
	default:
		addNode(tag)
		// return fmt.Errorf("unknown tag id%.4x-tag%.4x-%s offset 0x%.6x", tag.Id, tag.Tag, tag.Name, tag.DebugPos)
	}
	return nil
}
