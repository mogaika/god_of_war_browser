package dma

import (
	"fmt"
)

type DmaTag uint64

const (
	DMA_TAG_CNTS = 0x00 //T=QWC D=QWC+1 MADR => STADR
	DMA_TAG_REFE = 0x00 //T=ADDR then END
	DMA_TAG_CNT  = 0x01 //T=QWC D=QWC+1
	DMA_TAG_NEXT = 0x02 //T=QWC D=ADDR
	DMA_TAG_REF  = 0x03 //D=D+1 T=ADDR
	DMA_TAG_REFS = 0x04 //.. + stall ctrl
	DMA_TAG_CALL = 0x05 //T=QWC D=ADDR QWC+1 => ASR0
	DMA_TAG_RET  = 0x06 //T=QWC (ASR0 => D) if !ASR0 then END
	DMA_TAG_END  = 0x07 //T=QWC then END
)

var dmaTagIdToString = []string{
	"cnts/refe", "cnt", "next", "ref",
	"refs", "call", "ret", "end",
}

func (p DmaTag) QWC() uint16 {
	return uint16(p & 0xffff)
}

func (p DmaTag) ID() uint8 {
	return uint8((p >> 28) & 0x7)
}

func (p DmaTag) Addr() uint32 {
	return uint32((p >> 32) & 0x7FFFFFFF)
}

func (p DmaTag) IsSPR() bool {
	return (p>>63)&1 != 0
}

func (p DmaTag) IsIRQ() bool {
	return (p>>31)&1 != 0
}

func (p DmaTag) String() string {
	return fmt.Sprintf("DmaTag{ID:%-9s; Addr:0x%.4x; QWC:0x%.2x; SPR:%t; IRQ:%t}",
		dmaTagIdToString[p.ID()], p.Addr(), p.QWC(), p.IsSPR(), p.IsIRQ())
}

func NewTag(raw uint64) DmaTag {
	return DmaTag(raw)
}
