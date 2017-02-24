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
		dmaPackPos := m.pos + i*16
		dmaPack := dma.NewTag(binary.LittleEndian.Uint64(m.Data[dmaPackPos:]))
		fmt.Fprintf(m.debugOut, " << << << = %v\n", dmaPack)
		switch dmaPack.ID() {
		case dma.DMA_TAG_REF:
			dataStart := dmaPack.Addr() + m.relativeAddr
			dataEnd := dataStart + uint32(dmaPack.QWC()*0x10)
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
		//return nil, fmt.Errorf("Empty xyzw array, possibly incorrect data: %x", debugPos)
		return nil, nil
	}
	return nil, nil
}
