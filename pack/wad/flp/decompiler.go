package flp

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type Label struct {
	Offset int16
	Name   string
}

type DecompilerType int
type DecompilerOffset int

const (
	DECOMPILER_TYPE_INT int = iota
	DECOMPILER_TYPE_STRING
	DECOMPILER_TYPE_FLOAT32
	DECOMPILER_TYPE_BOOL
	DECOMPILER_TYPE_LABEL
	DECOMPILER_TYPE_ANY
)

type Stack struct {
	Stack    [32]*Variable
	Position int8
}

func (s *Stack) Push(v *Variable) {
	s.Stack[s.Position] = v
	s.Position += 1
}

func (s *Stack) Pop() *Variable {
	s.Position -= 1
	return s.Stack[s.Position]
}

type Value interface {
	String() string
}

type Variable struct {
	Name       string
	References int
	Value      Value
	// Type DecompilerType
}

func (r *Variable) Reference() *Variable {
	r.References += 1
	return r
}

func NewVariable(name string, v Value) *Variable {
	return &Variable{Name: name, Value: v}
}

func (v *Variable) String() string {
	if v.References == 1 {
		return v.Value.String()
	} else {
		return "@" + v.Name
	}
}

type ValueConstString struct{ Str string }

func (v *ValueConstString) String() string { return "\"" + v.Str + "\"" }

func NewValueConstString(s string) *ValueConstString { return &ValueConstString{Str: s} }

type ValueConstInt struct{ Int int32 }

func (v *ValueConstInt) String() string { return fmt.Sprintf("%d", v.Int) }

func NewValueConstInt(v int32) *ValueConstInt { return &ValueConstInt{Int: v} }

type ValueConstBool struct{ Bool bool }

func (v *ValueConstBool) String() string { return fmt.Sprintf("%t", v.Bool) }

func NewValueConstBool(b bool) *ValueConstBool { return &ValueConstBool{Bool: b} }

type Op interface {
	String() string
}

type OpBinary struct {
	Op  string
	Arg [2]Value
}

func (o *OpBinary) String() string {
	var toStr func(v Value) string
	toStr = func(v Value) string {
		brackets := false
		if _, ok := v.(*OpBinary); ok {
			brackets = true
		}
		if subOpVar, ok := v.(*Variable); ok {
			if subOpVar.References == 1 {
				brackets = true
				return toStr(subOpVar.Value)
			}
		}
		if brackets {
			return "(" + v.String() + ")"
		} else {
			return v.String()
		}
	}
	return toStr(o.Arg[0]) + " " + o.Op + " " + toStr(o.Arg[1])
}

func NewOpBinary(a1 Value, op string, a2 Value) *OpBinary {
	o := &OpBinary{Op: op}
	o.Arg[0] = a1
	o.Arg[1] = a2
	return o
}

type OpIfElse struct {
	Condition           *Variable
	LineTrue, LineFalse Op
}

func (o *OpIfElse) String() string {
	/*
			sT := o.LineTrue.String()
			sF := o.LineFalse.String()
		pushSpaces := func(s string) string {
			newS := ""
			for _, line := range strings.Split(s, "\n") {
				if line != "" {
					newS += "» " + s
				}
			}
			return newS
		}
		return "if (" + o.Condition.String() + ") {\n" + pushSpaces(sT) + "\n} else {\n" + pushSpaces(sF) + "\n}\n"
	*/
	return "if (" + o.Condition.String() + ")"
}

func (o *OpIfElse) StringNot() string {
	return "if not (" + o.Condition.String() + ")"
}

func NewOpIfElse(condition *Variable, t, f Op) *OpIfElse {
	return &OpIfElse{
		Condition: condition,
		LineTrue:  t, LineFalse: f,
	}
}

type OpAssign struct {
	Target *Variable
}

func (o *OpAssign) String() string {
	if o.Target.References != 1 {
		return o.Target.String() + " = " + o.Target.Value.String()
	} else {
		return ""
	}
}

func NewOpAssign(v *Variable) *OpAssign { return &OpAssign{Target: v} }

type OpJump struct {
	Target *Line
}

func (o *OpJump) String() string {
	block := &o.Target.d.blocks[o.Target.BlockId]
	for block.Forwarded {
		block = &block.d.blocks[block.BlockOutputs[0]]
	}
	return fmt.Sprintf("jump #label_%d", block.Id)
}

func NewOpJump(target *Line) *OpJump {
	return &OpJump{Target: target.MarkAsLabel()}
}

type OpMethod struct {
	Name       string
	Parameters []Value
}

func NewOpMethod(name string, params ...Value) *OpMethod {
	return &OpMethod{Name: name, Parameters: params}
}

func (om *OpMethod) String() string {
	params := make([]string, len(om.Parameters))
	for i, p := range om.Parameters {
		params[i] = p.String()
	}
	return om.Name + "(" + strings.Join(params, ", ") + ")"
}

type BlockId int

type Block struct {
	d            *Decompiler
	Id           BlockId
	FirstLineId  LineId
	LastLineId   LineId
	BlockInputs  []BlockId
	BlockOutputs []BlockId
	Forwarded    bool
	MergingIn    []BlockId
}

func (b *Block) connectAsOutput(otherBlock *Block) {
	b.BlockOutputs = append(b.BlockOutputs, otherBlock.Id)
	otherBlock.BlockInputs = append(otherBlock.BlockInputs, b.Id)
}

func (b *Block) connectAsOutputById(otherBlockId BlockId) {
	b.connectAsOutput(&b.d.blocks[otherBlockId])
}

func (b *Block) connectAsOutputByJump(op *OpJump) {
	b.connectAsOutputById(op.Target.BlockId)
}

func (b *Block) countMerghingBlocks() []BlockId {
	if b.MergingIn != nil {
		return b.MergingIn
	}

	defer func() {
		sort.Slice(b.MergingIn, func(i, j int) bool { return b.MergingIn[i] < b.MergingIn[j] })
	}()

	b.MergingIn = make([]BlockId, 0, 8)
	if len(b.BlockOutputs) == 1 {
		b.MergingIn = append(b.MergingIn, b.BlockOutputs[0])
		b.MergingIn = append(b.MergingIn, b.d.blocks[b.BlockOutputs[0]].countMerghingBlocks()...)
		return b.MergingIn
	}

	blockIdExistsInArray := func(b BlockId, array []BlockId) bool {
		for _, v := range array {
			if v == b {
				return true
			}
		}
		return false
	}

	if len(b.BlockOutputs) == 2 {
		b0Mergings := b.d.blocks[b.BlockOutputs[0]].countMerghingBlocks()
		b1Mergings := b.d.blocks[b.BlockOutputs[1]].countMerghingBlocks()

		if blockIdExistsInArray(b.BlockOutputs[0], b1Mergings) {
			b.MergingIn = append(b.MergingIn, b.BlockOutputs[0])
		}
		if blockIdExistsInArray(b.BlockOutputs[1], b0Mergings) {
			b.MergingIn = append(b.MergingIn, b.BlockOutputs[1])
		}

		checkArrayForMergings := func(a1, a2 []BlockId) {
			for _, candidate := range a1 {
				if blockIdExistsInArray(candidate, b.MergingIn) {
					continue
				}
				if blockIdExistsInArray(candidate, a2) {
					b.MergingIn = append(b.MergingIn, candidate)
				}
			}
		}

		checkArrayForMergings(b0Mergings, b1Mergings)
		checkArrayForMergings(b1Mergings, b0Mergings)
	}
	/*
		subMergings := make([][]BlockId, len(b.BlockOutputs))

		for i, outBlockId := range b.BlockOutputs {
			subMergings[i] = b.d.blocks[outBlockId].countMerghingBlocks()
		}

		for i, outMergings := range subMergings {
			for _, candidate := range outMergings {
				if blockIdExistsInArray(candidate, b.MergingIn) {
					continue
				}
				failed := false
				for j, outMergingsToCompare := range subMergings {
					if i == j {
						continue
					}
					if !blockIdExistsInArray(candidate, outMergingsToCompare) {
						failed = true
						break
					}
				}
				if !failed {
					b.MergingIn = append(b.MergingIn, candidate)
				}
			}
		}
	*/
	return b.MergingIn
}

func (b *Block) String(tabs string, scopeLastBlock BlockId) string {
	if b.Forwarded {
		return b.d.blocks[b.BlockOutputs[0]].String(tabs, scopeLastBlock)
	}

	s := ""

	/*
		s += tabs + fmt.Sprintf(":label_%d // offset 0x%.4x f %d l %d ins: %v outs: %v\n",
			b.Id, b.d.lines[b.FirstLineId].ScriptOpcode.Offset,
			b.FirstLineId, b.LastLineId,
			b.BlockInputs, b.BlockOutputs)
		s += tabs + fmt.Sprintf("//  merging in %v\n", b.MergingIn)
	*/

	for lineId := b.FirstLineId; lineId <= b.LastLineId; lineId++ {
		line := b.d.lines[lineId]
		//s += fmt.Sprintf("&#47;&#47; line %d op %+#v\n", lineId, line.Op)
		switch op := line.Op.(type) {
		case *OpJump:
		default:
			sline := line.Op.String()
			if sline != "" {
				s += tabs + sline
				if _, isIf := line.Op.(*OpIfElse); !isIf {
					s += ";"
				}
				s += tabs + "\n"
			}
		case *OpIfElse:
			trueBlockId := op.LineTrue.(*OpJump).Target.BlockId
			falseBlockId := op.LineFalse.(*OpJump).Target.BlockId
			trueBlock := &b.d.blocks[trueBlockId]
			falseBlock := &b.d.blocks[falseBlockId]

			if falseBlockId == scopeLastBlock {
				// TODO: add last block inverted special case
				if falseBlockId == BlockId(len(b.d.blocks)-1) {
					s += tabs + op.StringNot() + " {\n"
					s += tabs + "  return;\n"
					s += tabs + "}\n"

					s += trueBlock.String(tabs, b.MergingIn[0])
				} else {
					s += tabs + op.String() + " {\n"
					s += trueBlock.String(tabs+"  ", b.MergingIn[0])
					s += tabs + "}\n"
				}
			} else {
				findMergingBlock := func(b1, b2 *Block) BlockId {
					for _, candidate := range b1.MergingIn {
						for _, candidate2 := range b2.MergingIn {
							if candidate == candidate2 {
								return candidate
							}
						}
					}
					log.Panicf("Wasn't able to find merging block for %v:%v %v:%v",
						b1.Id, b1.MergingIn, b2.Id, b2.MergingIn)
					return -1
				}
				findFirstMergingBlock := func(b1, b2 *Block) BlockId {
					r1 := findMergingBlock(b1, b2)
					r2 := findMergingBlock(b1, b2)
					if r1 < r2 {
						return r1
					}
					return r2
				}

				mergingBlock := findFirstMergingBlock(trueBlock, falseBlock)

				s += tabs + op.String() + " {\n"
				s += trueBlock.String(tabs+"  ", mergingBlock)
				s += tabs + "} else {\n"
				s += falseBlock.String(tabs+"  ", mergingBlock)
				s += tabs + "}\n"

				if mergingBlock != BlockId(len(b.d.blocks)-1) {
					s += b.d.blocks[mergingBlock].String(tabs, scopeLastBlock)
				}
			}
		case nil:
		}
	}

	return s
}

type LineId int
type Line struct {
	d            *Decompiler
	Id           LineId
	Label        bool
	ScriptOpcode *ScriptOpcode
	Op           Op
	Stack        Stack
	BlockId      BlockId
}

func (l *Line) MarkAsLabel() *Line {
	l.Label = true
	return l
}

func (l *Line) NewOpAssignAndPush(v *Variable) *OpAssign {
	op := NewOpAssign(v)
	l.Op = op
	l.Stack.Push(v)
	return op
}

type Decompiler struct {
	s         *Script
	lines     []Line
	blocks    []Block
	opToLines map[*ScriptOpcode]LineId
}

func NewDecompiler(s *Script) *Decompiler {
	return &Decompiler{
		s: s,
	}
}

func (d *Decompiler) parseLines() {
	d.lines = make([]Line, 0, 16)
	d.opToLines = make(map[*ScriptOpcode]LineId)

	for _, op := range d.s.Opcodes {
		lineId := LineId(len(d.lines))
		d.lines = append(d.lines,
			Line{
				d:            d,
				Id:           lineId,
				ScriptOpcode: op,
			})
		d.opToLines[op] = lineId
	}
}

func (d *Decompiler) getLineForOffset(offset DecompilerOffset) int {
	for lineId := range d.lines {
		opOff := DecompilerOffset(d.lines[lineId].ScriptOpcode.Offset)
		if opOff == offset {
			return lineId
		} else if opOff > offset {
			break
		}
	}
	panic("not found lineid for such opcode")
}

func (d *Decompiler) scriptLabelToLine(sLabel string) *Line {
	/*
		for l := range d.s.labelToOffset {
			log.Printf("Searching label %q to == %q", l, sLabel)
		}
	*/
	off, ex := d.s.labelToOffset[sLabel]
	if !ex {
		log.Panicf("Offset not existing for label %q. Available: %v", sLabel, d.s.labelToOffset)
	}
	return &d.lines[d.getLineForOffset(DecompilerOffset(off))]
}

func (d *Decompiler) buildOps(lineId int, stack Stack) error {
	if lineId >= len(d.lines) {
		return nil
	}
	line := &d.lines[lineId]
	if line.Op != nil {
		return nil
	}

	line.Stack = stack
	/*
		log.Printf("Parsing %.4x: 0x%.2x %q  Stack %v",
			line.ScriptOpcode.Offset, line.ScriptOpcode.Code,
			line.ScriptOpcode.String(), line.Stack.Stack[:line.Stack.Position])
	*/
	sop := line.ScriptOpcode

	switch sop.Code {
	case 0x00:
		return nil
	case 0x06:
		line.Op = NewOpMethod("Play")
	case 0x07:
		line.Op = NewOpMethod("Stop")
	case 0xa: // float +
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_float_%d", sop.Offset),
			NewOpBinary(v2, "+", v1)))
	case 0xb: // float -
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_float_%d", sop.Offset),
			NewOpBinary(v2, "-", v1)))
	case 0xd: // float /
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_float_%d", sop.Offset),
			NewOpBinary(v2, "/", v1)))
	case 0xe: // float ==
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_bool_%d", sop.Offset),
			NewOpBinary(v2, "==", v1)))
	case 0x0f: // float <
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_bool_%d", sop.Offset),
			NewOpBinary(v2, "<", v1)))
	case 0x10: // binary AND
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_bool_%d", sop.Offset),
			NewOpBinary(v2, "&&", v1)))
	case 0x11: // binary OR
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_bool_%d", sop.Offset),
			NewOpBinary(v2, "||", v1)))
	case 0x12:
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_bool_%d", sop.Offset),
			NewOpMethod("bool", line.Stack.Pop().Reference())))
	case 0x13:
		// TODO: may be inverted like !strcmp
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_bool_%d", sop.Offset),
			NewOpMethod("strcmp", line.Stack.Pop().Reference(), line.Stack.Pop().Reference())))
	case 0x17:
		line.Op = NewOpMethod("unknown_017_discard", line.Stack.Pop().Reference())
	case 0x18:
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_%d", sop.Offset),
			NewOpMethod("round", line.Stack.Pop().Reference())))
	case 0x1c:
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_%d", sop.Offset),
			NewOpMethod("VFS::Get", line.Stack.Pop().Reference())))
	case 0x1d:
		v := line.Stack.Pop().Reference()
		t := line.Stack.Pop().Reference()
		line.Op = NewOpMethod("VFS::Set", t, v)
	case 0x20:
		line.Op = NewOpMethod("SetTarget", line.Stack.Pop().Reference())
	case 0x21: // string +
		v1 := line.Stack.Pop().Reference()
		v2 := line.Stack.Pop().Reference()
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_float_%d", sop.Offset),
			NewOpMethod("append", v2, v1)))
	case 0x34: // push current timer
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_timer_%d", sop.Offset),
			NewOpMethod("CurrentTime")))
	case 0x81:
		line.Op = NewOpMethod("GotoFrame", NewValueConstInt(int32(sop.Parameters[0].(uint16))))
	case 0x83:
		line.Op = NewOpMethod("Fs_queue_or_response result",
			NewValueConstString(sop.Parameters[0].(string)),
			NewValueConstString(sop.Parameters[1].(string)))
	case 0x8b:
		line.Op = NewOpMethod("SetTarget", NewValueConstString(sop.Parameters[0].(string)))
	case 0x8c:
		line.Op = NewOpMethod("GotoLabel", NewValueConstString(sop.Parameters[0].(string)))
	case 0x96: // push string
		line.NewOpAssignAndPush(NewVariable(fmt.Sprintf("var_str_%d", sop.Offset),
			NewValueConstString(sop.Parameters[0].(string))))
	case 0x99: // jump #label
		line.Op = NewOpJump(d.scriptLabelToLine(sop.Parameters[0].(string)))
	case 0x9d: // if pop == true { jump #label }
		condition := line.Stack.Pop().Reference()
		line.Op = NewOpIfElse(condition,
			NewOpJump(&d.lines[lineId+1]),
			NewOpJump(d.scriptLabelToLine(sop.Parameters[0].(string))),
		)
	case 0x9e: // call frame
		line.Op = NewOpMethod("CallFrame", line.Stack.Pop().Reference())
	case 0x9f: // goto expression
		line.Op = NewOpMethod("GotoExpression",
			line.Stack.Pop().Reference(),
			NewValueConstBool(sop.Parameters[0].(bool)))
	default:
		line.Op = NewOpMethod(fmt.Sprintf("UNKNOWN_CODE_%.4x_0x%.2x_%q st %v",
			sop.Offset, sop.Code, sop.String(), stack.Stack[:stack.Position]))
		return nil
	}

	return d.buildOps(lineId+1, line.Stack)
}

func (d *Decompiler) buildBlocks() {
	d.blocks = make([]Block, 0, 16)
	newBlock := func(l *Line) *Block {
		blockId := BlockId(len(d.blocks))
		d.blocks = append(d.blocks, Block{
			Id:           blockId,
			d:            d,
			FirstLineId:  l.Id,
			LastLineId:   l.Id,
			BlockInputs:  make([]BlockId, 0, 2),
			BlockOutputs: make([]BlockId, 0, 2),
		})
		return &d.blocks[blockId]
	}

	var block *Block
	for lineId := range d.lines {
		line := &d.lines[lineId]
		if block == nil || line.Label {
			block = newBlock(line)
		}
		line.BlockId = block.Id
		block.LastLineId = line.Id
	}
}

func (d *Decompiler) buildBlocksConnections() {
	for blockId := range d.blocks {
		block := &d.blocks[blockId]
		lastLine := &d.lines[block.LastLineId]

		switch op := lastLine.Op.(type) {
		case *OpIfElse:
			block.connectAsOutputByJump(op.LineTrue.(*OpJump))
			block.connectAsOutputByJump(op.LineFalse.(*OpJump))
		case *OpJump:
			block.connectAsOutputByJump(op)
		case nil:
		default:
			block.connectAsOutputById(block.Id + 1)
		}
	}
}

func (d *Decompiler) optimizeJumps() {
	for blockId := range d.blocks {
		block := &d.blocks[blockId]
		lastLine := &d.lines[block.LastLineId]

		if block.FirstLineId == block.LastLineId {
			// if only one line in block
			if _, isJump := lastLine.Op.(*OpJump); isJump {
				// and it's line is jump
				// just reconnect all input blocks to this jump out block
				block.Forwarded = true

				for _, inBlockId := range block.BlockInputs {
					inBlock := d.blocks[inBlockId]
					for i, outBlockId := range inBlock.BlockOutputs {
						if outBlockId == block.Id {
							inBlock.BlockOutputs[i] = block.BlockOutputs[0]
						}
					}
				}
			}
		}
	}
}

func (d *Decompiler) countMergingBlocks() {
	for blockId := range d.blocks {
		d.blocks[blockId].countMerghingBlocks()
	}
}

func (d *Decompiler) Decompile() string {
	d.parseLines()
	if err := d.buildOps(0, Stack{}); err != nil {
		return fmt.Sprintf("ERROR: Failed to get build OPS: %v", err)
	}

	d.buildBlocks()
	d.buildBlocksConnections()
	d.countMergingBlocks()

	return d.blocks[0].String("", BlockId(len(d.blocks)-1))
}
