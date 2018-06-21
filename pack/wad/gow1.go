package wad

import "fmt"

const (
	TAG_GOW1_ENTITY_COUNT     = 24
	TAG_GOW1_SERVER_INSTANCE  = 30
	TAG_GOW1_FILE_GROUP_START = 40
	TAG_GOW1_FILE_GROUP_END   = 50

	TAG_GOW1_FILE_MC_DATA  = 110
	TAG_GOW1_FILE_MC_ICON  = 111
	TAG_GOW1_FILE_RAW_DATA = 112
	TAG_GOW1_TWK_INSTANCE  = 113
	TAG_GOW1_TWK_OBJECT    = 114

	TAG_GOW1_RSRCS = 500

	TAG_GOW1_DATA_START1 = 666
	TAG_GOW1_DATA_START2 = 80
	TAG_GOW1_DATA_START3 = 777

	TAG_GOW1_HEADER_START = 888
	TAG_GOW1_HEADER_POP   = 999
)

func (w *Wad) gow1parseTag(tag *Tag, currentNode *NodeId, newGroupTag *bool, addNode func(tag *Tag) *Node) error {
	switch tag.Tag {
	case TAG_GOW1_SERVER_INSTANCE: // file data packet
		// Tell file server (server determined by first uint16)
		// that new file is loaded
		// if name start with space, then name ignored (unnamed)
		// overwrite previous instance with same name
		n := addNode(tag)
		if *newGroupTag {
			*newGroupTag = false
			*currentNode = n.Id
		}
	case TAG_GOW1_FILE_GROUP_START: // file data group start
		*newGroupTag = true // same realization in game
	case TAG_GOW1_FILE_GROUP_END: // file data group end
		// Tell server about group ended
		if !*newGroupTag {
			// theres been some nodes and we change currentNode
			if *currentNode == NODE_INVALID {
				return fmt.Errorf("Trying to end not started group id%d-%s", tag.Id, tag.Name)
			}
			*currentNode = w.GetNodeById(*currentNode).Parent
		} else {
			*newGroupTag = false
		}
	case TAG_GOW1_ENTITY_COUNT: // entity count
		// Game also add empty named node to nodedirectory?
	case TAG_GOW1_FILE_MC_DATA: // MC_DATA   < R_PERM.WAD
		// Just add node to nodedirectory
		addNode(tag)
	case TAG_GOW1_FILE_MC_ICON: // MC_ICON   < R_PERM.WAD
		// Like 0x006e, but also store size of data
		addNode(tag)
	case TAG_GOW1_FILE_RAW_DATA: // MSH_BDepoly6Shape
		// Add node to nodedirectory only if
		// another node with this name not exists
		addNode(tag)
	case TAG_GOW1_TWK_INSTANCE: // TWK_Cloth_195
		// Tweaks affect VFS of game
		// AI logics, animation
		// exec twk asap?
		addNode(tag)
	case TAG_GOW1_TWK_OBJECT: // TWK_CombatFile_328
		// Affect VFS too
		// store twk in mem?
		addNode(tag)
	case TAG_GOW1_RSRCS: // RSRCS
		// probably affect WadReader
		// (internally transformed to R_RSRCS)
		addNode(tag)
	case TAG_GOW1_DATA_START1, TAG_GOW1_DATA_START2, TAG_GOW1_DATA_START3: // file data start
		// PopBatchServerStack of server from first uint16
		addNode(tag)
	case TAG_GOW1_HEADER_START: // file header start
		// create new memory namespace and push to memorystack
		// create new nodedirectory and push to nodestack
		// data loading init
		addNode(tag)
	case TAG_GOW1_HEADER_POP: // file header pop heap
		// data loading structs cleanup
		addNode(tag)
	default:
		return fmt.Errorf("unknown tag id%.4x-tag%.4x-%s offset 0x%.6x", tag.Id, tag.Tag, tag.Name, tag.DebugPos)
	}
	return nil
}
