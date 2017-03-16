package mesh

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/ps2/dma"
	"github.com/mogaika/god_of_war_browser/ps2/vif"
)

/*
mem:
0x0000:0x*    - vertex meta

0x0136:0x0005 - texture info

0x155 <<<< next block
*/

var unpackBuffersBases = []uint32{0, 0x155, 0x2ab, 0x400}

const GSFixed12Point4Delimeter = 16.0
const GSFixed12Point4Delimeter1000 = 4096.0

type MeshDataStream struct {
	Data         []byte
	PacketsCount uint32
	blocks       []stBlock
	pos          uint32
	relativeAddr uint32
	state        *MeshParserState
	debugOut     io.Writer
}

func NewMeshDataStream(data []byte, packetsCount uint32, pos uint32, chainRelativeAdd uint32, debugOut io.Writer) *MeshDataStream {
	return &MeshDataStream{
		Data:         data,
		PacketsCount: packetsCount,
		pos:          pos,
		debugOut:     debugOut,
		relativeAddr: chainRelativeAdd,
		blocks:       make([]stBlock, 0),
	}
}

func (m *MeshDataStream) Log(format string, pos uint32, args ...interface{}) {
	f := fmt.Sprintf("      %.6x %s\n", pos, format)
	fmt.Fprintf(m.debugOut, f, args...)
}

func (m *MeshDataStream) Blocks() []stBlock {
	if m.blocks == nil || len(m.blocks) == 0 {
		return nil
	} else {
		return m.blocks
	}
}

func (m *MeshDataStream) flushState(debugPos uint32) error {
	if m.state != nil {
		block, err := m.state.ToBlock(debugPos, m.debugOut)
		if err != nil {
			return err
		}
		if block != nil {
			m.blocks = append(m.blocks, *block)
		}
	}
	m.state = &MeshParserState{}
	return nil
}

func (m *MeshDataStream) ParseVifStreamLegacy(vif []byte, debug_off uint32, debugOut io.Writer) (error, []*stBlock) {

	result := make([]*stBlock, 0)

	var block_data_xyzw []byte = nil
	var block_data_rgba []byte = nil
	var block_data_uv []byte = nil
	block_data_uv_width := 0
	var block_data_norm []byte = nil
	var block_data_vertex_meta []byte = nil

	pos := uint32(0)
	spaces := "     "
	exit := false
	flush := false

	for iCommandInBlock := 0; !exit; iCommandInBlock++ {
		pos = ((pos + 3) / 4) * 4
		if pos >= uint32(len(vif)) {
			break
		}

		pk_cmd := vif[pos+3]
		pk_num := vif[pos+2]
		pk_dat2 := vif[pos+1]
		pk_dat1 := vif[pos]

		tagpos := pos
		pos += 4

		if pk_cmd >= 0x60 { // if unpack command
			components := ((pk_cmd >> 2) & 0x3) + 1
			bwidth := pk_cmd & 0x3
			widthmap := []uint32{32, 16, 8, 4} // 4 = r5g5b5a1
			width := widthmap[bwidth]

			blocksize := uint32(components) * ((width * uint32(pk_num)) / 8)

			signed := ((pk_dat2&(1<<6))>>6)^1 != 0
			address := (pk_dat2&(1<<7))>>7 != 0

			target := uint32(pk_dat1) | (uint32(pk_dat2&3) << 8)

			handledBy := ""

			switch width {
			case 32:
				if signed {
					switch components {
					case 4: // joints and format info all time after data (i think)
						flush = true
						handledBy = "meta"
						for i := byte(0); i < pk_num; i++ {
							bp := pos + uint32(i)*0x10
							fmt.Fprintf(debugOut, "%s -  %.6x = %.4x %.4x %.4x %.4x  %.4x %.4x %.4x %.4x\n", spaces, debug_off+bp,
								binary.LittleEndian.Uint16(vif[bp:bp+2]), binary.LittleEndian.Uint16(vif[bp+2:bp+4]),
								binary.LittleEndian.Uint16(vif[bp+4:bp+6]), binary.LittleEndian.Uint16(vif[bp+6:bp+8]),
								binary.LittleEndian.Uint16(vif[bp+8:bp+10]), binary.LittleEndian.Uint16(vif[bp+10:bp+12]),
								binary.LittleEndian.Uint16(vif[bp+12:bp+14]), binary.LittleEndian.Uint16(vif[bp+14:bp+16]))
						}
						switch target {
						case 0x000, 0x155, 0x2ab:
							block_data_vertex_meta = vif[pos : pos+blocksize]
							handledBy = "vmta"
						}
					case 2:
						handledBy = " uv4"
						if block_data_uv == nil {
							block_data_uv = vif[pos : pos+blocksize]
							handledBy = " uv2"
							block_data_uv_width = 4
						} else {
							return fmt.Errorf("UV already present. What is this: %.6x ?", tagpos+debug_off), nil
						}
					}
				}
			case 16:
				if signed {
					switch components {
					case 4:
						if block_data_xyzw == nil {
							block_data_xyzw = vif[pos : pos+blocksize]
							handledBy = "xyzw"
						} else {
							return fmt.Errorf("XYZW already present. What is this: %.6x ?", tagpos+debug_off), nil
						}
					case 2:
						if block_data_uv == nil {
							block_data_uv = vif[pos : pos+blocksize]
							handledBy = " uv2"
							block_data_uv_width = 2
						} else {
							return fmt.Errorf("UV already present. What is this: %.6x ?", tagpos+debug_off), nil
						}
					}
				}
			case 8:
				if signed {
					switch components {
					case 3:
						if block_data_norm == nil {
							block_data_norm = vif[pos : pos+blocksize]
							handledBy = "norm"
						} else {
							return fmt.Errorf("NORM already present. What is this: %.6x ?", tagpos+debug_off), nil
						}
					}
				} else {
					switch components {
					case 4:
						if block_data_rgba == nil {
							block_data_rgba = vif[pos : pos+blocksize]
							handledBy = "rgba"
						} else {
							return fmt.Errorf("RGBA already present. What is this: %.6x ?", tagpos+debug_off), nil
						}
					}
				}
			}

			if handledBy == "" {
				return fmt.Errorf("Block %.6x (cmd %.2x; %d bit; %d components; %d elements; sign %t; addr %t; target: %.3x; size: %.6x) not handled",
					tagpos+debug_off, pk_cmd, width, components, pk_num, signed, address, target, blocksize), nil
			} else {
				fmt.Fprintf(debugOut, "%s %.6x vif unpack [%s]: %.2x elements: %.2x components: %d type: %.2d target: %.3x sign: %t addr: %t size: %.6x\n",
					spaces, debug_off+tagpos, handledBy, pk_cmd, pk_num, components, width, target, signed, address, blocksize)
			}

			pos += blocksize
		} else {
			switch pk_cmd {
			case 0:
				fmt.Fprintf(debugOut, "%s %.6x nop\n", spaces, debug_off+tagpos)
			case 01:
				fmt.Fprintf(debugOut, "%s %.6x Stcycl wl=%.2x cl=%.2x\n", spaces, debug_off+tagpos, pk_dat2, pk_dat1)
			case 05:
				cmode := " pos "
				/*	 Decompression modes
				Normal = 0,
				OffsetDecompression, // would conflict with vif code
				Difference
				*/
				switch pk_dat1 {
				case 1:
					cmode = "[pos]"
				case 2:
					cmode = "[cur]"
				}
				fmt.Fprintf(debugOut, "%s %.6x Stmod  mode=%s (%d)\n", spaces, debug_off+tagpos, cmode, pk_dat1)
			case 0x14:
				fmt.Fprintf(debugOut, "%s %.6x Mscall proc command\n", spaces, debug_off+tagpos)
				flush = true
			case 0x30:
				fmt.Fprintf(debugOut, "%s %.6x Strow  proc command\n", spaces, debug_off+tagpos)
				pos += 0x10
			default:
				return fmt.Errorf("Unknown %.6x VIF command: %.2x:%.2x data: %.2x:%.2x",
					debug_off+tagpos, pk_cmd, pk_num, pk_dat1, pk_dat2), nil
			}
		}

		if flush || exit {
			flush = false

			// if we collect some data
			if block_data_xyzw != nil {
				currentBlock := &stBlock{}
				currentBlock.DebugPos = tagpos

				countTrias := len(block_data_xyzw) / 8
				currentBlock.Trias.X = make([]float32, countTrias)
				currentBlock.Trias.Y = make([]float32, countTrias)
				currentBlock.Trias.Z = make([]float32, countTrias)
				currentBlock.Trias.Skip = make([]bool, countTrias)
				for i := range currentBlock.Trias.X {
					bp := i * 8
					currentBlock.Trias.X[i] = float32(int16(binary.LittleEndian.Uint16(block_data_xyzw[bp:bp+2]))) / GSFixed12Point4Delimeter
					currentBlock.Trias.Y[i] = float32(int16(binary.LittleEndian.Uint16(block_data_xyzw[bp+2:bp+4]))) / GSFixed12Point4Delimeter
					currentBlock.Trias.Z[i] = float32(int16(binary.LittleEndian.Uint16(block_data_xyzw[bp+4:bp+6]))) / GSFixed12Point4Delimeter
					currentBlock.Trias.Skip[i] = block_data_xyzw[bp+7]&0x80 != 0
				}

				if block_data_uv != nil {
					switch block_data_uv_width {
					case 2:
						uvCount := len(block_data_uv) / 4
						currentBlock.Uvs.U = make([]float32, uvCount)
						currentBlock.Uvs.V = make([]float32, uvCount)
						for i := range currentBlock.Uvs.U {
							bp := i * 4
							currentBlock.Uvs.U[i] = float32(int16(binary.LittleEndian.Uint16(block_data_uv[bp:bp+2]))) / GSFixed12Point4Delimeter1000
							currentBlock.Uvs.V[i] = float32(int16(binary.LittleEndian.Uint16(block_data_uv[bp+2:bp+4]))) / GSFixed12Point4Delimeter1000
						}
					case 4:
						uvCount := len(block_data_uv) / 8
						currentBlock.Uvs.U = make([]float32, uvCount)
						currentBlock.Uvs.V = make([]float32, uvCount)
						for i := range currentBlock.Uvs.U {
							bp := i * 8
							currentBlock.Uvs.U[i] = float32(int32(binary.LittleEndian.Uint32(block_data_uv[bp:bp+4]))) / GSFixed12Point4Delimeter1000
							currentBlock.Uvs.V[i] = float32(int32(binary.LittleEndian.Uint32(block_data_uv[bp+4:bp+8]))) / GSFixed12Point4Delimeter1000
						}
					}
				}

				if block_data_norm != nil {
					normcnt := len(block_data_norm) / 3
					currentBlock.Norms.X = make([]float32, normcnt)
					currentBlock.Norms.Y = make([]float32, normcnt)
					currentBlock.Norms.Z = make([]float32, normcnt)
					for i := range currentBlock.Norms.X {
						bp := i * 3
						currentBlock.Norms.X[i] = float32(int8(block_data_norm[bp])) / 100.0
						currentBlock.Norms.Y[i] = float32(int8(block_data_norm[bp+1])) / 100.0
						currentBlock.Norms.Z[i] = float32(int8(block_data_norm[bp+2])) / 100.0
					}
				}

				if block_data_rgba != nil {
					rgbacnt := len(block_data_rgba) / 4
					currentBlock.Blend.R = make([]uint16, rgbacnt)
					currentBlock.Blend.G = make([]uint16, rgbacnt)
					currentBlock.Blend.B = make([]uint16, rgbacnt)
					currentBlock.Blend.A = make([]uint16, rgbacnt)
					for i := range currentBlock.Blend.R {
						bp := i * 4
						currentBlock.Blend.R[i] = uint16(block_data_rgba[bp])
						currentBlock.Blend.G[i] = uint16(block_data_rgba[bp+1])
						currentBlock.Blend.B[i] = uint16(block_data_rgba[bp+2])
						currentBlock.Blend.A[i] = uint16(block_data_rgba[bp+3])
					}
				}

				if block_data_vertex_meta != nil {
					blocks := len(block_data_vertex_meta) / 16
					vertexes := len(currentBlock.Trias.X)

					currentBlock.Joints = make([]uint16, vertexes)

					vertnum := 0
					for i := 0; i < blocks; i++ {
						block := block_data_vertex_meta[i*16 : i*16+16]

						block_verts := int(block[0])

						for j := 0; j < block_verts; j++ {
							currentBlock.Joints[vertnum+j] = uint16(block[12]) | (uint16((block[13] / 4)) << 8)
						}

						vertnum += block_verts

						if block[1]&0x80 != 0 {
							if i != blocks-1 {
								return fmt.Errorf("Block count != blocks: %v <= %v", blocks, i), nil
							}
						}
					}
					if vertnum != vertexes {
						return fmt.Errorf("Vertnum != vertexes count: %v <= %v", vertnum, vertexes), nil
					}
				}

				result = append(result, currentBlock)

				fmt.Fprintf(debugOut, "%s = Flush xyzw:%t, rgba:%t, uv:%t, norm:%t\n", spaces,
					block_data_xyzw != nil, block_data_rgba != nil,
					block_data_uv != nil, block_data_norm != nil)

				block_data_norm = nil
				block_data_rgba = nil
				block_data_xyzw = nil
				block_data_uv = nil

			}
		}
	}
	return nil, result
}

func (m *MeshDataStream) ParseVifStream(data []byte, debugPos uint32) error {
	pos := uint32(0)
	for {
		pos = ((pos + 3) / 4) * 4
		if pos >= uint32(len(data)) {
			break
		}
		tagPos := pos
		vifCode := vif.NewCode(binary.LittleEndian.Uint32(data[pos:]))

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
			if m.state == nil {
				m.state = &MeshParserState{Buffer: vifBufferBase}
			} else if vifBufferBase != m.state.Buffer {
				if err := m.flushState(debugPos + tagPos); err != nil {
					return err
				}
				m.state.Buffer = vifBufferBase
			}
			handledBy := ""

			defer func() {
				if r := recover(); r != nil {
					m.Log("!! !! panic on unpack [%s]: 0x%.2x elements: 0x%.2x components: %d width: %.2d target: 0x%.3x sign: %t tops: %t size: %.6x", debugPos+tagPos,
						handledBy, vifCode.Cmd(), vifCode.Num(), vifComponents, vifWidth, vifTarget, vifIsSigned, vifUseTops, vifBlockSize)
					panic(r)
				}
			}()

			vifBlock := data[pos : pos+vifBlockSize]

			errorAlreadyPresent := func(handler string) error {
				m.Log("++> unpack [%s]: 0x%.2x elements: 0x%.2x components: %d width: %.2d target: 0x%.3x sign: %t tops: %t size: %.6x", debugPos+tagPos,
					handledBy, vifCode.Cmd(), vifCode.Num(), vifComponents, vifWidth, vifTarget, vifIsSigned, vifUseTops, vifBlockSize)
				return fmt.Errorf("%s already present. What is this: %.6x ?", handler, tagPos+debugPos)
			}

			switch vifWidth {
			case 32:
				if vifIsSigned {
					switch vifComponents {
					case 4: // joints and format info all time after data (i think)
						switch vifTarget {
						case 0x000, 0x155, 0x2ab:
							if m.state.VertexMeta != nil {
								return errorAlreadyPresent("Vertex Meta")
							}
							m.state.VertexMeta = vifBlock
							handledBy = "vmta"
						default:
							if m.state.Meta != nil {
								return errorAlreadyPresent("Meta")
							}
							m.state.Meta = vifBlock
							handledBy = "meta"
						}
					case 2:
						handledBy = " uv4"
						if m.state.UV == nil {
							m.state.UV = vifBlock
							handledBy = " uv2"
							m.state.UVWidth = 4
						} else {
							return errorAlreadyPresent("UV")
						}
					}
				}
			case 16:
				if vifIsSigned {
					switch vifComponents {
					case 4:
						if m.state.XYZW == nil {
							m.state.XYZW = vifBlock
							handledBy = "xyzw"
						} else {
							return errorAlreadyPresent("XYZW")
						}
					case 2:
						if m.state.UV == nil {
							m.state.UV = vifBlock
							handledBy = " uv2"
							m.state.UVWidth = 2
						} else {
							return errorAlreadyPresent("UV")
						}
					}
				}
			case 8:
				if vifIsSigned {
					switch vifComponents {
					case 3:
						if m.state.Norm == nil {
							m.state.Norm = vifBlock
							handledBy = "norm"
						} else {
							return errorAlreadyPresent("Norm")
						}
					}
				} else {
					switch vifComponents {
					case 4:
						//if m.state.RGBA == nil {
						if m.state.RGBA != nil {
							m.Log(" --overwrite of rgba data. Compare=%d", debugPos+tagPos, bytes.Compare(m.state.RGBA, vifBlock))
						}
						m.state.RGBA = vifBlock
						handledBy = "rgba"
						//} else {
						//	return errorAlreadyPresent("RGBA")
						//}
					}
				}
			}

			m.Log("+ unpack [%s]: 0x%.2x elements: 0x%.2x components: %d width: %.2d target: 0x%.3x sign: %t tops: %t size: %.6x", debugPos+tagPos,
				handledBy, vifCode.Cmd(), vifCode.Num(), vifComponents, vifWidth, vifTarget, vifIsSigned, vifUseTops, vifBlockSize)
			if handledBy == "" {
				return fmt.Errorf("Block %.6x (cmd 0x%.2x; %d bit; %d components; %d elements; sign %t; tops %t; target: %.3x; size: %.6x) not handled",
					tagPos+debugPos, vifCode.Cmd(), vifWidth, vifComponents, vifCode.Num(), vifIsSigned, vifUseTops, vifTarget, vifBlockSize)
			}
			pos += vifBlockSize
		} else {
			m.Log("# vif %v", tagPos+debugPos, vifCode)
			switch vifCode.Cmd() {
			case vif.VIF_CMD_MSCAL:
				if err := m.flushState(debugPos + pos); err != nil {
					return err
				}
			case vif.VIF_CMD_STROW:
				pos += 0x10
			}
		}
	}

	return nil
}

func (m *MeshDataStream) ParsePackets() error {
	for i := uint32(0); i < m.PacketsCount; i++ {
		//		packet: 0 pos: 0x000130 rows: 0x0112 end: 0x001250

		dmaPackPos := m.pos + i*16
		dmaPack := dma.NewTag(binary.LittleEndian.Uint64(m.Data[dmaPackPos:]))

		fmt.Fprintf(m.debugOut, "    packet: %d pos: 0x%.6x rows: 0x%.4x end: 0x%.6x\n",
			i, dmaPack.Addr()+m.relativeAddr, dmaPack.QWC(), dmaPack.Addr()+m.relativeAddr+uint32(dmaPack.QWC()*16))
		fmt.Fprintf(m.debugOut, " << << << = %v\n", dmaPack)
		switch dmaPack.ID() {
		case dma.DMA_TAG_REF:
			dataStart := dmaPack.Addr() + m.relativeAddr
			dataEnd := dataStart + dmaPack.QWC()*0x10
			m.Log("vif pack start calc: 0x%.6x + 0x%.6x = 0x%.6x => 0x%.6x", dmaPackPos, dmaPack.Addr(), m.relativeAddr, dataStart, dataEnd)
			if err := m.ParseVifStream(m.Data[dataStart:dataEnd], dataStart); err != nil {
				return err
			}
		case dma.DMA_TAG_RET:
			if dmaPack.QWC() != 0 {
				return fmt.Errorf("Not support dma_tag_ret with qwc != 0 (%d)", dmaPack.QWC())
			}
			if i != m.PacketsCount-1 {
				return fmt.Errorf("dma_tag_ret not in end of stream (%d != %d)", i, m.PacketsCount-1)
			} else {
				m.Log("dma_tag_ret in end of stream", dmaPackPos)
			}
		default:
			return fmt.Errorf("Unknown dma packet %v in mesh stream", dmaPack)
		}
	}
	m.flushState(m.relativeAddr)
	return nil
}

type MeshParserState struct {
	XYZW       []byte
	RGBA       []byte
	UV         []byte
	UVWidth    int
	Norm       []byte
	Meta       []byte
	VertexMeta []byte
	Buffer     int
}

func (state *MeshParserState) ToBlock(debugPos uint32, debugOut io.Writer) (*stBlock, error) {
	if state.XYZW != nil {
		currentBlock := &stBlock{HasTransparentBlending: false}
		currentBlock.DebugPos = debugPos

		countTrias := len(state.XYZW) / 8
		currentBlock.Trias.X = make([]float32, countTrias)
		currentBlock.Trias.Y = make([]float32, countTrias)
		currentBlock.Trias.Z = make([]float32, countTrias)
		currentBlock.Trias.Skip = make([]bool, countTrias)
		for i := range currentBlock.Trias.X {
			bp := i * 8
			currentBlock.Trias.X[i] = float32(int16(binary.LittleEndian.Uint16(state.XYZW[bp:bp+2]))) / GSFixed12Point4Delimeter
			currentBlock.Trias.Y[i] = float32(int16(binary.LittleEndian.Uint16(state.XYZW[bp+2:bp+4]))) / GSFixed12Point4Delimeter
			currentBlock.Trias.Z[i] = float32(int16(binary.LittleEndian.Uint16(state.XYZW[bp+4:bp+6]))) / GSFixed12Point4Delimeter
			currentBlock.Trias.Skip[i] = state.XYZW[bp+7]&0x80 != 0
		}

		if state.UV != nil {
			switch state.UVWidth {
			case 2:
				uvCount := len(state.UV) / 4
				currentBlock.Uvs.U = make([]float32, uvCount)
				currentBlock.Uvs.V = make([]float32, uvCount)
				for i := range currentBlock.Uvs.U {
					bp := i * 4
					currentBlock.Uvs.U[i] = float32(int16(binary.LittleEndian.Uint16(state.UV[bp:bp+2]))) / GSFixed12Point4Delimeter1000
					currentBlock.Uvs.V[i] = float32(int16(binary.LittleEndian.Uint16(state.UV[bp+2:bp+4]))) / GSFixed12Point4Delimeter1000
				}
			case 4:
				uvCount := len(state.UV) / 8
				currentBlock.Uvs.U = make([]float32, uvCount)
				currentBlock.Uvs.V = make([]float32, uvCount)
				for i := range currentBlock.Uvs.U {
					bp := i * 8
					currentBlock.Uvs.U[i] = float32(int32(binary.LittleEndian.Uint32(state.UV[bp:bp+4]))) / GSFixed12Point4Delimeter1000
					currentBlock.Uvs.V[i] = float32(int32(binary.LittleEndian.Uint32(state.UV[bp+4:bp+8]))) / GSFixed12Point4Delimeter1000
				}
			}
		}

		if state.Norm != nil {
			normcnt := len(state.Norm) / 3
			currentBlock.Norms.X = make([]float32, normcnt)
			currentBlock.Norms.Y = make([]float32, normcnt)
			currentBlock.Norms.Z = make([]float32, normcnt)
			for i := range currentBlock.Norms.X {
				bp := i * 3
				currentBlock.Norms.X[i] = float32(int8(state.Norm[bp])) / 100.0
				currentBlock.Norms.Y[i] = float32(int8(state.Norm[bp+1])) / 100.0
				currentBlock.Norms.Z[i] = float32(int8(state.Norm[bp+2])) / 100.0
			}
		}

		if state.RGBA != nil {
			rgbacnt := len(state.RGBA) / 4
			currentBlock.Blend.R = make([]uint16, rgbacnt)
			currentBlock.Blend.G = make([]uint16, rgbacnt)
			currentBlock.Blend.B = make([]uint16, rgbacnt)
			currentBlock.Blend.A = make([]uint16, rgbacnt)
			for i := range currentBlock.Blend.R {
				bp := i * 4
				currentBlock.Blend.R[i] = uint16(state.RGBA[bp])
				currentBlock.Blend.G[i] = uint16(state.RGBA[bp+1])
				currentBlock.Blend.B[i] = uint16(state.RGBA[bp+2])
				currentBlock.Blend.A[i] = uint16(state.RGBA[bp+3])
			}
			for _, a := range currentBlock.Blend.A {
				if a < 0x80 {
					currentBlock.HasTransparentBlending = true
					break
				}
			}
		}

		if state.VertexMeta != nil {
			blocks := len(state.VertexMeta) / 16
			vertexes := len(currentBlock.Trias.X)

			currentBlock.Joints = make([]uint16, vertexes)

			vertnum := 0
			for i := 0; i < blocks; i++ {
				block := state.VertexMeta[i*16 : i*16+16]

				block_verts := int(block[0])

				for j := 0; j < block_verts; j++ {
					currentBlock.Joints[vertnum+j] = uint16(block[13] >> 4)
				}

				vertnum += block_verts

				if block[1]&0x80 != 0 {
					if i != blocks-1 {
						return nil, fmt.Errorf("Block count != blocks: %v <= %v", blocks, i)
					}
				}
			}
			if vertnum != vertexes {
				return nil, fmt.Errorf("Vertnum != vertexes count: %v <= %v", vertnum, vertexes)
			}
		}

		fmt.Fprintf(debugOut, "    = Flush xyzw:%t, rgba:%t, uv:%t, norm:%t, vmeta:%t (%d)\n",
			state.XYZW != nil, state.RGBA != nil, state.UV != nil,
			state.Norm != nil, state.VertexMeta != nil, len(currentBlock.Trias.X))
		return currentBlock, nil
	} else {
		if state.UV != nil || state.Norm != nil || state.VertexMeta != nil || state.RGBA != nil {
			return nil, fmt.Errorf("Empty xyzw array, possibly incorrect data: %x. State: %+#v", debugPos)
		}
		return nil, nil
	}
	return nil, nil
}
