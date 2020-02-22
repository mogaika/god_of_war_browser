package mesh

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/ps2/dma"
	"github.com/mogaika/god_of_war_browser/ps2/vif"
	"github.com/mogaika/god_of_war_browser/utils"
)

var unpackBuffersBases = []uint32{0, 0x155, 0x2ab}

const GSFixedPoint8 = 16.0
const GSFixedPoint24 = 4096.0

type MeshParserStream struct {
	Data                []byte
	Offset              uint32
	Packets             []Packet
	Object              *Object
	Log                 *utils.Logger
	state               *MeshParserState
	lastPacketDataStart uint32
	lastPacketDataEnd   uint32
}

type MeshParserState struct {
	XYZW       []byte
	RGBA       []byte
	UV         []byte
	UVWidth    int
	Norm       []byte
	Boundaries []byte
	VertexMeta []byte
	Buffer     int
}

func NewMeshParserStream(allb []byte, object *Object, packetOffset uint32, exlog *utils.Logger) *MeshParserStream {
	return &MeshParserStream{
		Data:    allb,
		Object:  object,
		Offset:  packetOffset,
		Log:     exlog,
		Packets: make([]Packet, 0),
	}
}

func (ms *MeshParserStream) flushState() error {
	if ms.state != nil {
		packet, err := ms.state.ToPacket(ms.Log, ms.lastPacketDataStart, ms.Object)
		if err != nil {
			return err
		}
		if packet != nil {
			ms.Packets = append(ms.Packets, *packet)
		}
	}
	ms.state = nil
	return nil
}

func (ms *MeshParserStream) ParsePackets() error {
	for i := uint32(0); i < ms.Object.DmaTagsCountPerPacket; i++ {
		dmaPackPos := ms.Offset + i*0x10
		dmaPack := dma.NewTag(binary.LittleEndian.Uint64(ms.Data[dmaPackPos:]))

		ms.Log.Printf("           -  dma offset: 0x%.8x packet: %d pos: 0x%.6x rows: 0x%.4x end: 0x%.6x", dmaPackPos,
			i, dmaPack.Addr()+ms.Object.Offset, dmaPack.QWC(), dmaPack.Addr()+ms.Object.Offset+uint32(dmaPack.QWC()*16))
		ms.Log.Printf("          | %v", dmaPack)
		switch dmaPack.ID() {
		case dma.DMA_TAG_REF:
			if err := ms.ParseVif(dmaPackPos+0x8, dmaPackPos+0x10); err != nil {
				return fmt.Errorf("Error when parsing vif stream triggered by injected dma_tag_ref: %v", err)
			}

			packetDataStart := dmaPack.Addr() + ms.Object.Offset
			packetDataEnd := packetDataStart + dmaPack.QWC()*0x10

			ms.Log.Printf("            -vif pack start: 0x%.6x + 0x%.6x = 0x%.6x => 0x%.6x",
				ms.Offset, dmaPack.Addr(), packetDataStart, packetDataEnd)

			if err := ms.ParseVif(packetDataStart, packetDataEnd); err != nil {
				return fmt.Errorf("Error when parsing vif stream triggered by dma_tag_ref: %v", err)
			}
		case dma.DMA_TAG_RET:
			if dmaPack.QWC() != 0 {
				return fmt.Errorf("Not support dma_tag_ret with qwc != 0 (%d)", dmaPack.QWC())
			}

			if err := ms.ParseVif(dmaPackPos+0x8, dmaPackPos+0x10); err != nil {
				return fmt.Errorf("Error when parsing vif stream triggered by injected dma_tag_ref: %v", err)
			}

			if i != ms.Object.DmaTagsCountPerPacket-1 {
				return fmt.Errorf("dma_tag_ret not in end of stream (%d != %d)", i, ms.Object.DmaTagsCountPerPacket-1)
			} else {
				ms.Log.Printf("             << dma_tag_ret at 0x%.8x >>", dmaPackPos)
			}
		default:
			return fmt.Errorf("Unknown dma packet %v in mesh stream at 0x%.8x i = 0x%.2x < 0x%.2x", dmaPack, dmaPackPos, i, ms.Object.DmaTagsCountPerPacket)
		}
	}
	if ms.state != nil {
		return fmt.Errorf("Missed state end")
	}
	return nil
}

func (state *MeshParserState) ToPacket(exlog *utils.Logger, debugPos uint32, ooo *Object) (*Packet, error) {
	if state.XYZW == nil {
		if state.UV != nil || state.Norm != nil || state.VertexMeta != nil || state.RGBA != nil {
			return nil, fmt.Errorf("Empty xyzw array, possibly incorrect data: 0x%x. State: %+#v", debugPos, state)
		}
		return nil, nil
	}

	packet := &Packet{HasTransparentBlending: false}
	packet.Offset = debugPos

	countTrias := len(state.XYZW) / 8
	packet.Trias.X = make([]float32, countTrias)
	packet.Trias.Y = make([]float32, countTrias)
	packet.Trias.Z = make([]float32, countTrias)
	packet.Trias.Skip = make([]bool, countTrias)
	packet.Trias.Weight = make([]float32, countTrias)
	for i := range packet.Trias.X {
		bp := i * 8
		packet.Trias.X[i] = float32(int16(binary.LittleEndian.Uint16(state.XYZW[bp:bp+2]))) / GSFixedPoint8
		packet.Trias.Y[i] = float32(int16(binary.LittleEndian.Uint16(state.XYZW[bp+2:bp+4]))) / GSFixedPoint8
		packet.Trias.Z[i] = float32(int16(binary.LittleEndian.Uint16(state.XYZW[bp+4:bp+6]))) / GSFixedPoint8

		flags := binary.LittleEndian.Uint16(state.XYZW[bp+6 : bp+8])
		packet.Trias.Skip[i] = flags&0x8000 != 0
		packet.Trias.Weight[i] = float32(int(flags&0x7fff)) / GSFixedPoint24
		if packet.Trias.Weight[i] > 1.0 || packet.Trias.Weight[i] < 0.0 {
			return nil, fmt.Errorf("Invalid trias wieght: %v (0x%x)", packet.Trias.Weight[i], flags)
		}
	}

	if state.UV != nil {
		switch state.UVWidth {
		case 2:
			uvCount := len(state.UV) / 4
			packet.Uvs.U = make([]float32, uvCount)
			packet.Uvs.V = make([]float32, uvCount)
			for i := range packet.Uvs.U {
				bp := i * 4
				packet.Uvs.U[i] = float32(int16(binary.LittleEndian.Uint16(state.UV[bp:bp+2]))) / GSFixedPoint24
				packet.Uvs.V[i] = float32(int16(binary.LittleEndian.Uint16(state.UV[bp+2:bp+4]))) / GSFixedPoint24
			}
		case 4:
			uvCount := len(state.UV) / 8
			packet.Uvs.U = make([]float32, uvCount)
			packet.Uvs.V = make([]float32, uvCount)
			for i := range packet.Uvs.U {
				bp := i * 8
				packet.Uvs.U[i] = float32(int32(binary.LittleEndian.Uint32(state.UV[bp:bp+4]))) / GSFixedPoint24
				packet.Uvs.V[i] = float32(int32(binary.LittleEndian.Uint32(state.UV[bp+4:bp+8]))) / GSFixedPoint24
			}
		}
	}

	if state.Norm != nil {
		normcnt := len(state.Norm) / 3
		packet.Norms.X = make([]float32, normcnt)
		packet.Norms.Y = make([]float32, normcnt)
		packet.Norms.Z = make([]float32, normcnt)
		for i := range packet.Norms.X {
			bp := i * 3
			packet.Norms.X[i] = float32(int8(state.Norm[bp])) / 127.0
			packet.Norms.Y[i] = float32(int8(state.Norm[bp+1])) / 127.0
			packet.Norms.Z[i] = float32(int8(state.Norm[bp+2])) / 127.0
		}
	}

	if state.RGBA != nil {
		rgbacnt := len(state.RGBA) / 4
		packet.Blend.R = make([]uint16, rgbacnt)
		packet.Blend.G = make([]uint16, rgbacnt)
		packet.Blend.B = make([]uint16, rgbacnt)
		packet.Blend.A = make([]uint16, rgbacnt)
		for i := range packet.Blend.R {
			bp := i * 4
			packet.Blend.R[i] = uint16(state.RGBA[bp])
			packet.Blend.G[i] = uint16(state.RGBA[bp+1])
			packet.Blend.B[i] = uint16(state.RGBA[bp+2])
			packet.Blend.A[i] = uint16(state.RGBA[bp+3])
		}
		for _, a := range packet.Blend.A {
			if a < 0x80 {
				packet.HasTransparentBlending = true
				break
			}
		}
	}

	if state.Boundaries != nil {
		for i := range packet.Boundaries {
			packet.Boundaries[i] = math.Float32frombits(binary.LittleEndian.Uint32(state.Boundaries[i*4 : i*4+4]))
		}
	}

	if state.VertexMeta != nil {
		blocks := len(state.VertexMeta) / 0x10
		packet.VertexMeta = state.VertexMeta
		vertexes := len(packet.Trias.X)

		packet.Joints = make([]uint16, vertexes)
		packet.Joints2 = make([]uint16, vertexes)

		vertnum := 0
		for i := 0; i < blocks; i++ {
			block := state.VertexMeta[i*16 : i*16+16]
			block_verts := int(block[0])

			debugColor := func(r, g, b uint16) {
				for j := 0; j < block_verts; j++ {
					packet.Blend.R[vertnum+j] = r
					packet.Blend.G[vertnum+j] = g
					packet.Blend.B[vertnum+j] = b
				}
			}
			_ = debugColor

			// block[0] = affected vertex count
			// block[1] = 0x80 if last block, else 0
			// block[2:4] = 0x00
			// block[4:8] = 4bit fields [
			//        0:4 - count of matbit flags
			//        4:8 - 1 if 12:16 == 7 or 12:16 == 3, elsewhere 0
			//       8:12 - 0
			//      12:16 - depends on block id. either {4,0...} either {7,6,0...}. or if in middle then {0...,3,2,0...}
			//      16:20 - 0xe - have texture (have rgb?), 0x6 - no texture, 0x2 - no texture (have rgb?), shadow layer (no rgb?)
			//      20:24 - 0x2 || 0x4;
			//      24:28 - 0
			//      28:32 - count of matbit flags
			// block[8:12] = 4bit matflags, stacked in particular order
			//        0x2 - if have texture
			//        0xf - if npc (do not depend on animations availability) / accept lighting? / lit?
			//        0x1 - if not gui
			//        0x5 - always (end of flags mark?)
			// block[12] = jointid1 * 4
			// block[13] = jointid2 * 16
			// block[14] = 0
			// block[15] = 0x80 if jointids == 0, but not limited??? / flags ?
			if true { //verification
				for iBlock, v := range block {
					ok := true
					lo := v & 0xf
					hi := v >> 4
					switch iBlock {
					case 1:
						ok = v == 0 || v == 0x80
					case 2:
						ok = v == 0
					case 3:
						ok = v == 0
					case 4:
						ok = (lo != 0 && lo <= 4) &&
							(hi == 0 || hi == 1)
					case 5:
						ok = (lo == 0) &&
							(hi == 0 || hi == 2 || hi == 3 || hi == 4 || hi == 6 || hi == 7)
						/*
							r := uint16(0)
							g := r
							b := r
							if hi&1 == 1 {
								r = 0xff
							}
							if hi&2 == 2 {
								g = 0xff
							}
							if hi&4 == 4 {
								b = 0xff
							}
							debugColor(r, g, b)
						*/
					case 6:
						/*
							if hi == 0x2 {
								debugColor(0xff, 0, 0)
							} else {
								debugColor(0, 0xff, 0)
							}
						*/
						ok = (lo == 0x2 || lo == 0x6 || lo == 0xe) &&
							(hi == 0x2 || hi == 0x4)
					case 7:
						ok = (lo == 0) &&
							(hi == block[4]&0xf)
					case 12:
						ok = v%4 == 0
					case 14:
						ok = v == 0
					case 15:
						ok = v == 0 || v == 0x80
					}
					if ok == false {
						return nil, fmt.Errorf("Incorrect block[%d]: 0x%x", iBlock, block[iBlock])
					}
				}
			}

			/*
				switch ooo.NextFreeVUBufferId {
				case 0:
					debugColor(0xff, 0, 0)
				case 1:
					debugColor(0, 0xff, 0)
				case 2:
					debugColor(0, 0, 0xff)
				default:
					debugColor(0xff, 0xff, 0xff)
				}
			*/

			for j := 0; j < block_verts; j++ {
				packet.Joints[vertnum+j] = uint16(block[13] >> 4)
				if packet.Joints2 != nil {
					packet.Joints2[vertnum+j] = uint16(block[12] >> 2)
				}
			}

			vertnum += block_verts
			if block[1] != 0 {
				if i != blocks-1 {
					return nil, fmt.Errorf("Block count != blocks: %v <= %v", blocks, i)
				}
			}
		}
		if vertnum != vertexes {
			return nil, fmt.Errorf("Vertnum != vertexes count: %v <= %v", vertnum, vertexes)
		}
	}

	exlog.Printf("    = Flush xyzw:%t, rgba:%t, uv:%t, norm:%t, vmeta:%t (%d)",
		state.XYZW != nil, state.RGBA != nil, state.UV != nil,
		state.Norm != nil, state.VertexMeta != nil, len(packet.Trias.X))

	atoStr := func(a []byte) string {
		u16 := func(barr []byte, id int) uint16 {
			return binary.LittleEndian.Uint16(barr[id*2 : id*2+2])
		}
		u32 := func(barr []byte, id int) uint32 {
			return binary.LittleEndian.Uint32(barr[id*4 : id*4+4])
		}
		f32 := func(barr []byte, id int) float32 {
			return math.Float32frombits(u32(barr, id))
		}
		return fmt.Sprintf(" %.4x %.4x  %.4x %.4x   %.4x %.4x  %.4x %.4x   |  %.8x %.8x %.8x %.8x |  %f  %f  %f  %f",
			u16(a, 0), u16(a, 1), u16(a, 2), u16(a, 3), u16(a, 4), u16(a, 5), u16(a, 6), u16(a, 7),
			u32(a, 0), u32(a, 1), u32(a, 2), u32(a, 3),
			f32(a, 0), f32(a, 1), f32(a, 2), f32(a, 3),
		)
	}

	if state.VertexMeta != nil {
		exlog.Printf("         Vertex Meta:")
		for i := 0; i < len(packet.VertexMeta)/16; i++ {
			exlog.Printf("  %s", atoStr(packet.VertexMeta[i*16:i*16]))
		}
	}

	if state.Boundaries != nil {
		exlog.Printf("         Boundaries: %v", packet.Boundaries)
	}

	return packet, nil
}

func (ms *MeshParserStream) ParseVif(packetDataStart, packetDataEnd uint32) error {
	ms.lastPacketDataStart = packetDataStart
	ms.lastPacketDataEnd = packetDataEnd

	data := ms.Data[packetDataStart:packetDataEnd]
	pos := uint32(0)
	for {
		pos = ((pos + 3) / 4) * 4
		if pos >= uint32(len(data)) {
			break
		}
		tagPos := pos
		rawVifCode := binary.LittleEndian.Uint32(data[pos:])
		vifCode := vif.NewCode(rawVifCode)

		pos += 4
		if vifCode.Cmd() > 0x60 {
			vifComponents := ((vifCode.Cmd() >> 2) & 0x3) + 1
			vifWidth := []uint32{32, 16, 8, 4}[vifCode.Cmd()&0x3]

			vifBlockSize := uint32(vifComponents) * ((vifWidth * uint32(vifCode.Num())) / 8)

			vifIsSigned := (vifCode.Imm()>>14)&1 == 0
			vifUseTops := (vifCode.Imm()>>15)&1 != 0
			vifTarget := uint32(vifCode.Imm() & 0x3ff)

			vifBufferBase := 1
			for _, base := range unpackBuffersBases {
				if vifTarget >= base {
					vifBufferBase++
				} else {
					break
				}
			}
			if ms.state == nil {
				ms.state = &MeshParserState{Buffer: vifBufferBase}
			} else if vifBufferBase != ms.state.Buffer {
				return fmt.Errorf("Unflushed prev state")
			}
			handledBy := ""

			defer func() {
				if r := recover(); r != nil {
					ms.Log.Printf("[%.4s] !! !! panic on unpack: 0x%.2x elements: 0x%.2x components: %d width: %.2d target: 0x%.3x sign: %t tops: %t size: %.6x",
						handledBy, vifCode.Cmd(), vifCode.Num(), vifComponents, vifWidth, vifTarget, vifIsSigned, vifUseTops, vifBlockSize)
					panic(r)
				}
			}()

			vifBlock := data[pos : pos+vifBlockSize]

			errorAlreadyPresent := func(handler string) error {
				ms.Log.Printf("already present [%.4s]++> unpack: 0x%.2x elements: 0x%.2x components: %d width: %.2d target: 0x%.3x sign: %t tops: %t size: %.6x",
					handledBy, vifCode.Cmd(), vifCode.Num(), vifComponents, vifWidth, vifTarget, vifIsSigned, vifUseTops, vifBlockSize)
				return fmt.Errorf("%s already present. What is this: %.6x ?", handler, tagPos+ms.lastPacketDataStart)
			}

			var detectingErr error = nil

			switch vifWidth {
			case 32:
				if vifIsSigned {
					switch vifComponents {
					case 4: // joints and format info all time after data (i think)
						switch vifTarget {
						case 0x000, 0x155, 0x2ab:
							if ms.state.VertexMeta != nil {
								detectingErr = errorAlreadyPresent("Vertex Meta")
							}
							ms.state.VertexMeta = vifBlock
							handledBy = "vmta"
						default:
							if ms.state.Boundaries != nil {
								detectingErr = errorAlreadyPresent("Boundaries")
							}
							ms.state.Boundaries = vifBlock
							handledBy = "bndr"
						}
					case 2:
						handledBy = " uv4"
						if ms.state.UV == nil {
							ms.state.UV = vifBlock
							ms.state.UVWidth = 4
						} else {
							detectingErr = errorAlreadyPresent("UV")
						}
					}
				}
			case 16:
				if vifIsSigned {
					switch vifComponents {
					case 4:
						if ms.state.XYZW == nil {
							ms.state.XYZW = vifBlock
							handledBy = "xyzw"
						} else {
							detectingErr = errorAlreadyPresent("XYZW")
						}
					case 2:
						if ms.state.UV == nil {
							ms.state.UV = vifBlock
							handledBy = " uv2"
							ms.state.UVWidth = 2
						} else {
							detectingErr = errorAlreadyPresent("UV")
						}
					}
				}
			case 8:
				switch vifComponents {
				case 3:
					if vifIsSigned {
						if ms.state.Norm == nil {
							ms.state.Norm = vifBlock
							handledBy = "norm"
						} else {
							detectingErr = errorAlreadyPresent("Norm")
						}
					}
				case 4:
					if ms.state.RGBA == nil {
						ms.state.RGBA = vifBlock
						handledBy = "rgba"
					} else {
						// game developers bug, for multi-layer models they include rgba twice
						ms.Log.Printf("- - - - - - - duplicate of rgba data")
						handledBy = "rgbA"
					}
					ms.state.RGBA = vifBlock
				}
			}

			ms.Log.Printf("[%.4s] + unpack: 0x%.2x cmd: 0x%.2x elements: 0x%.2x components: %d width: %.2d target: 0x%.3x sign: %t tops: %t size: %.6x",
				handledBy, ms.lastPacketDataStart+tagPos, vifCode.Cmd(), vifCode.Num(), vifComponents, vifWidth, vifTarget, vifIsSigned, vifUseTops, vifBlockSize)

			ms.Log.Printf("   __RAW: %s", utils.SDump(vifBlock))

			if detectingErr != nil || handledBy == "" {
				ms.Log.Printf("ERROR: %v", detectingErr)
				ms.Log.Printf("RAW: %s", utils.SDump(vifBlock))
				ms.Log.Printf("Current state: %s", utils.SDump(ms.state))
				if detectingErr != nil {
					return detectingErr
				} else {
					return fmt.Errorf("Block 0x%.6x (cmd 0x%.2x; %d bit; %d components; %d elements; sign %t; tops %t; target: %.3x; size: %.6x) not handled",
						tagPos+ms.lastPacketDataStart, vifCode.Cmd(), vifWidth, vifComponents, vifCode.Num(), vifIsSigned, vifUseTops, vifTarget, vifBlockSize)
				}
			}
			pos += vifBlockSize
		} else {
			ms.Log.Printf("# vif %v", vifCode)
			switch vifCode.Cmd() {
			case vif.VIF_CMD_NOP:
			case vif.VIF_CMD_STCYCL:
			case vif.VIF_CMD_MSCAL:
				if err := ms.flushState(); err != nil {
					return err
				}
			case vif.VIF_CMD_STROW:
				ms.Log.Printf("     - VIF STROW %v", utils.SDump(data[pos:pos+10]))
				pos += 0x10
			default:
				ms.Log.Printf("     - unknown VIF: %v", vifCode)
			}
		}
	}

	return nil
}
