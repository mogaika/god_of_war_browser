package vif

import "fmt"

type VifCode uint32

const (
	VIF_CMD_NOP      = 0x00 // No Operation
	VIF_CMD_STCYCL   = 0x01 // Sets CYCLE register
	VIF_CMD_OFFSET   = 0x02 // Sets OFFSET register (VIF1)
	VIF_CMD_BASE     = 0x03 // Sets BASE register (VIF1)
	VIF_CMD_ITOP     = 0x04 // Sets ITOPS register
	VIF_CMD_STMOD    = 0x05 // Sets MODE register
	VIF_CMD_MSKPATH3 = 0x06 // Mask GIF transfer (VIF1)
	VIF_CMD_MARK     = 0x07 // Sets Mark register
	VIF_CMD_FLUSHE   = 0x10 // Wait for end of microprogram
	VIF_CMD_FLUSH    = 0x11 // Wait for end of microprogram & Path 1/2 GIF xfer (VIF1)
	VIF_CMD_FLUSHA   = 0x13 // Wait for end of microprogram & all Path GIF xfer (VIF1)
	VIF_CMD_MSCAL    = 0x14 // Activate microprogram
	VIF_CMD_MSCNT    = 0x17 // Execute microrprogram continuously
	VIF_CMD_MSCALF   = 0x15 // Activate microprogram (VIF1)
	VIF_CMD_STMASK   = 0x20 // Sets MASK register
	VIF_CMD_STROW    = 0x30 // Sets ROW register
	VIF_CMD_STCOL    = 0x31 // Sets COL register
	VIF_CMD_MPG      = 0x4A // Load microprogram
	VIF_CMD_DIRECT   = 0x50 // Transfer data to GIF (VIF1)
	VIF_CMD_DIRECTHL = 0x51 // Transfer data to GIF but stall for Path 3 IMAGE mode (VIF1)
)

func (v VifCode) Cmd() uint8 {
	return uint8((v >> 24) & 0xff)
}

func (v VifCode) Num() uint8 {
	return uint8((v >> 16) & 0xff)
}

func (v VifCode) Imm() uint16 {
	return uint16(v & 0xffff)
}

func (v VifCode) IsIRQ() bool {
	return (v>>31)&1 != 0
}

func (v VifCode) String() string {
	return fmt.Sprintf("VifCode{Cmd:0x%.2x; Num:0x%.2x; Imm:0x%.4x; IRQ:%t}",
		v.Cmd(), v.Num(), v.Imm(), v.IsIRQ())
}

func NewCode(raw uint32) VifCode {
	return VifCode(raw)
}
