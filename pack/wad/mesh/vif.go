package mesh

import (
	"encoding/binary"
	"fmt"
	"io"
)

type stUV struct {
	U, V float32
}

type stNorm struct {
	X, Y, Z float32
}

type stRGBA struct {
	R, G, B, A uint8
}

type stXYZ struct {
	X, Y, Z float32
	Skip    bool
}

type stBlock struct {
	Uvs      []stUV
	Trias    []stXYZ
	Norms    []stNorm
	Blend    []stRGBA
	Joints   []uint16
	DebugPos uint32
}

// GS use 12:4 fixed point format
// 1 << 4 = 16
const GSFixed12Point4Delimeter = 16.0
const GSFixed12Point4Delimeter1000 = 4096.0

func VifRead1(vif []byte, debug_off uint32, debugOut io.Writer) (error, []*stBlock) {
	/*
		What game send on vif:

		Stcycl wl=1 cl=1/2/3/4

		One of array:
		[ xyzw4_16i ] -
			only position (GUI)
		[ rgba4_08u , xyzw4_16i ] -
			color and position (GUI + Effects)
		[ uv2_16i , xyzw4_16i ] -
			texture coords and position (simple model)
		[ uv2_16i , rgba4_08u , xyzw4_16i ] -
			texture coords + blend for hand-shaded models + position
		[ uv2_16i , norm2_16i , rgba4_08u , xyzw4_16i ] -
			texture coords + normal vector + blend color + position for hard materials

		Stcycl wl=1 cl=1

		Command vars:
		[ xyzw4_32i ] -
			paket refrence (verticles count, joint assign, joint types).
			used stable targets: 000, 155, 2ab
		[ xyzw4_32i ] -
			material refrence ? (diffuse/ambient colors, alpha)?

		Mscall (if not last sequence in packet) - process data

		Anyway position all time xyzw4_16i and last in sequence
	*/

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

				currentBlock.Trias = make([]stXYZ, len(block_data_xyzw)/8)
				for i := range currentBlock.Trias {
					bp := i * 8
					t := &currentBlock.Trias[i]
					t.X = float32(int16(binary.LittleEndian.Uint16(block_data_xyzw[bp:bp+2]))) / GSFixed12Point4Delimeter
					t.Y = float32(int16(binary.LittleEndian.Uint16(block_data_xyzw[bp+2:bp+4]))) / GSFixed12Point4Delimeter
					t.Z = float32(int16(binary.LittleEndian.Uint16(block_data_xyzw[bp+4:bp+6]))) / GSFixed12Point4Delimeter
					t.Skip = block_data_xyzw[bp+7]&0x80 != 0
				}

				if block_data_uv != nil {
					switch block_data_uv_width {
					case 2:
						currentBlock.Uvs = make([]stUV, len(block_data_uv)/4)
						for i := range currentBlock.Trias {
							bp := i * 4
							u := &currentBlock.Uvs[i]
							u.U = float32(int16(binary.LittleEndian.Uint16(block_data_uv[bp:bp+2]))) / GSFixed12Point4Delimeter1000
							u.V = float32(int16(binary.LittleEndian.Uint16(block_data_uv[bp+2:bp+4]))) / GSFixed12Point4Delimeter1000
						}
					case 4:
						currentBlock.Uvs = make([]stUV, len(block_data_uv)/8)
						for i := range currentBlock.Trias {
							bp := i * 8
							u := &currentBlock.Uvs[i]
							u.U = float32(int32(binary.LittleEndian.Uint32(block_data_uv[bp:bp+4]))) / GSFixed12Point4Delimeter1000
							u.V = float32(int32(binary.LittleEndian.Uint32(block_data_uv[bp+4:bp+8]))) / GSFixed12Point4Delimeter1000
						}
					}
				}

				if block_data_norm != nil {
					currentBlock.Norms = make([]stNorm, len(block_data_norm)/3)
					for i := range currentBlock.Norms {
						bp := i * 3
						n := &currentBlock.Norms[i]
						n.X = float32(int8(block_data_norm[bp])) / 100.0
						n.Y = float32(int8(block_data_norm[bp+1])) / 100.0
						n.Z = float32(int8(block_data_norm[bp+2])) / 100.0
					}
				}

				if block_data_rgba != nil {
					currentBlock.Blend = make([]stRGBA, len(block_data_rgba)/4)
					for i := range currentBlock.Blend {
						bp := i * 4
						c := &currentBlock.Blend[i]
						c.R = block_data_rgba[bp]
						c.G = block_data_rgba[bp+1]
						c.B = block_data_rgba[bp+2]
						c.A = block_data_rgba[bp+3]
					}
				}

				if block_data_vertex_meta != nil {
					blocks := len(block_data_vertex_meta) / 16
					vertexes := len(currentBlock.Trias)

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
