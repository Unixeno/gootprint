package frame

import (
	"bytes"
)

type ForFrame struct {
	*baseFrame
	varName string
}

func NewForFrame(path string) *ForFrame {
	return &ForFrame{baseFrame: NewBaseFrame(path)}
}

func (frame *ForFrame) GenBeginning(genEnv *baseEnv, content []byte) []byte {
	return content
}

func (frame *ForFrame) GenEnding(genEnv *baseEnv, content []byte) []byte {
	frame.varName = genEnv.genPointVarName()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genEnv.genCollect(frame.varName))
	buf.Write(content)
	return buf.Bytes()
}

func (frame *ForFrame) GenEnv(genEnv *baseEnv) []byte {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(genEnv.genPoint(frame.varName, frame.getStdPath()))
	return buffer.Bytes()
}
