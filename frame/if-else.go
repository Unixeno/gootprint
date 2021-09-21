package frame

import (
	"bytes"
)

type IfElseFrame struct {
	*baseFrame
	varName string
}

func NewIfElseFrame(path string) *IfElseFrame {
	return &IfElseFrame{
		baseFrame: NewBaseFrame(path),
	}
}

func (frame *IfElseFrame) GenBeginning(genEnv *baseEnv, content []byte) []byte {
	return content
}

func (frame *IfElseFrame) GenEnding(genEnv *baseEnv, content []byte) []byte {
	frame.varName = genEnv.genPointVarName()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genEnv.genCollect(frame.varName))
	buf.Write(content)
	return buf.Bytes()
}

func (frame *IfElseFrame) GenEnv(genEnv *baseEnv) []byte {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(genEnv.genPoint(frame.varName, frame.getStdPath()))
	return buffer.Bytes()
}
