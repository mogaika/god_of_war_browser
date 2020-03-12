package utils

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
)

type BufStack struct {
	parent         *BufStack
	childs         []*BufStack
	buf            []byte
	relativeOffset int
	absoluteOffset int
	size           int
	pos            int
	kind           string
	name           string
}

func NewBufStack(kind string, b []byte) *BufStack {
	return &BufStack{
		buf:  b,
		size: len(b),
		kind: kind,
	}
}

func (bs *BufStack) addChild(childBs *BufStack) {
	if bs.childs == nil {
		bs.childs = make([]*BufStack, 1)
		bs.childs[0] = childBs
	} else {
		index := sort.Search(len(bs.childs), func(i int) bool {
			return bs.childs[i].relativeOffset > childBs.relativeOffset
		})
		bs.childs = append(bs.childs, childBs)
		copy(bs.childs[index+1:], bs.childs[index:])
		bs.childs[index] = childBs
	}
}

func (bs *BufStack) SubBuf(kind string, offset int) *BufStack {
	childBs := &BufStack{
		parent:         bs,
		relativeOffset: offset,
		absoluteOffset: bs.absoluteOffset + offset,
		kind:           kind,
		buf:            bs.buf[offset:],
	}
	bs.addChild(childBs)
	return childBs
}

func (bs *BufStack) SubBufFollowing(kind string) *BufStack {
	if bs.size == 0 {
		panic(fmt.Sprintf("buffer %v size == 0", bs))
	}
	return bs.parent.SubBuf(kind, bs.relativeOffset+bs.size)
}

func (bs *BufStack) SetName(name string) *BufStack {
	bs.name = name
	return bs
}

func (bs *BufStack) SetSize(size int) *BufStack {
	bs.size = size
	return bs
}

func (bs *BufStack) Expand() *BufStack {
	if bs.parent.size == 0 {
		panic(fmt.Sprintf("buffer %v parent %v size == 0", bs, bs.parent))
	}
	bs.size = bs.parent.size - bs.relativeOffset
	return bs
}

func (bs *BufStack) Name() string {
	return bs.name
}

func (bs *BufStack) Size() int {
	return bs.size
}

func (bs *BufStack) Kind() string {
	return bs.kind
}

func (bs *BufStack) Parent() *BufStack {
	return bs.parent
}

func (bs *BufStack) RelativeOffset() int {
	return bs.relativeOffset
}

func (bs *BufStack) String() string {
	return fmt.Sprintf("buf<%v>(%v)[o:0x%x,s:0x%x,ao:0x%x,ae:0x%x]",
		bs.kind, bs.name, bs.relativeOffset, bs.size, bs.absoluteOffset, bs.absoluteOffset+bs.size)
}

func (bs *BufStack) StringChain() string {
	s := bs.String()
	if bs.parent != nil {
		s += fmt.Sprintf("::%s", bs.parent.String())
	}
	return s
}

func (bs *BufStack) Error() string {
	return bs.StringChain()
}

func (bs *BufStack) stringTree(pad int) string {
	sPad := ""
	for i := 0; i < pad; i++ {
		sPad += ".  "
	}
	s := sPad + bs.String() + "\n"
	pos := 0
	if bs.childs != nil {
		for i, child := range bs.childs {
			if pos >= 0 && child.relativeOffset > pos {
				s += fmt.Sprintf("%s.  gap [o:0x%x,s:0x%x,ao:0x%x,ae:0x%x]\n",
					sPad, pos, child.relativeOffset-pos, bs.absoluteOffset+pos, child.absoluteOffset)
			}
			s += child.stringTree(pad + 1)
			if child.size != 0 {
				pos = child.relativeOffset + child.size
			} else {
				pos = -1
			}
			if child.size > 0 {
				end := child.relativeOffset + child.size
				if i == len(bs.childs)-1 {
					if bs.size > 0 && end > bs.size {
						s += fmt.Sprintf("%s. [OVERGROW]\n", sPad)
					}
				} else {
					if end > bs.childs[i+1].relativeOffset {
						s += fmt.Sprintf("%s. [OVERLAP]\n", sPad)
					}
				}
			}
		}
	}
	return s
}

func (bs *BufStack) StringTree() string {
	return bs.stringTree(0)
}

func (bs *BufStack) Raw() []byte {
	raw := bs.buf[:]
	if bs.size != 0 {
		raw = raw[:bs.size]
	}
	return raw
}

func (bs *BufStack) Pos() int {
	return bs.pos
}

func (bs *BufStack) Read(amount int) []byte {
	oldPos := bs.pos
	bs.pos += amount
	return bs.buf[oldPos:bs.pos]
}

func (bs *BufStack) Skip(amount int) {
	bs.pos += amount
	if bs.size != 0 && bs.pos > bs.size {
		panic("skipped over buf")
	}
}

func (bs *BufStack) ReadLU64() uint64 {
	return binary.LittleEndian.Uint64(bs.Read(8))
}

func (bs *BufStack) ReadLU32() uint32 {
	return binary.LittleEndian.Uint32(bs.Read(4))
}

func (bs *BufStack) ReadLU16() uint16 {
	return binary.LittleEndian.Uint16(bs.Read(2))
}

func (bs *BufStack) ReadBU64() uint64 {
	return binary.BigEndian.Uint64(bs.Read(8))
}

func (bs *BufStack) ReadBU32() uint32 {
	return binary.BigEndian.Uint32(bs.Read(4))
}

func (bs *BufStack) ReadBU16() uint16 {
	return binary.BigEndian.Uint16(bs.Read(2))
}

func (bs *BufStack) ReadByte() byte {
	return bs.Read(1)[0]
}

func (bs *BufStack) ReadLF() float32 {
	return math.Float32frombits(bs.ReadLU32())
}

func (bs *BufStack) ReadBF() float32 {
	return math.Float32frombits(bs.ReadBU32())
}

func (bs *BufStack) ReadStringBuffer(size int) string {
	return BytesToString(bs.Read(size))
}

func (bs *BufStack) ReadZString(limit int) string {
	l := 0
	for i := 0; ; i++ {
		if i == limit {
			l = i
			break
		}
		if bs.buf[bs.pos+i] == 0 {
			l = i + 1
			break
		}
	}

	s := BytesToString(bs.buf[bs.pos : bs.pos+l])
	bs.pos += l
	return s
}

func (bs *BufStack) VerifySize(pos int) {
	if pos != bs.size {
		panic(fmt.Sprintf("Mismatch sizes: %v != %v", pos, bs.size))
	}
	if bs.size > len(bs.buf) {
		panic(fmt.Sprintf("Overgrown buffer: %v > %v", bs.size, len(bs.buf)))
	}
}

func (bs *BufStack) LU64(off int) uint64 {
	return binary.LittleEndian.Uint64(bs.buf[off:])
}

func (bs *BufStack) LU32(off int) uint32 {
	return binary.LittleEndian.Uint32(bs.buf[off:])
}

func (bs *BufStack) LU16(off int) uint16 {
	return binary.LittleEndian.Uint16(bs.buf[off:])
}

func (bs *BufStack) BU64(off int) uint64 {
	return binary.BigEndian.Uint64(bs.buf[off:])
}

func (bs *BufStack) BU32(off int) uint32 {
	return binary.BigEndian.Uint32(bs.buf[off:])
}

func (bs *BufStack) BU16(off int) uint16 {
	return binary.BigEndian.Uint16(bs.buf[off:])
}

func (bs *BufStack) Byte(off int) byte {
	return bs.buf[off]
}

func (bs *BufStack) LF(off int) float32 {
	return math.Float32frombits(bs.LU32(off))
}

func (bs *BufStack) BF(off int) float32 {
	return math.Float32frombits(bs.BU32(off))
}
