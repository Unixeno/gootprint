package frame

import (
	"bytes"
)

type FuncFrame struct {
	*baseFrame
}

func NewFuncFrame(path string) *FuncFrame {
	return &FuncFrame{NewBaseFrame(path)}
}

func (frame *FuncFrame) GenBeginning(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(content)
	buf.WriteString(genLineCodeWithStringArg("Collect", "call "+frame.path))
	return buf.Bytes()
}

func (frame *FuncFrame) GenEnding(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genLineCodeWithStringArg("Collect", frame.path))
	buf.Write(content)
	return buf.Bytes()
}
