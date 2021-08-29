package frame

import (
	"bytes"
)

type IfElseFrame struct {
	*baseFrame
}

func NewIfElseFrame(path string) *IfElseFrame {
	return &IfElseFrame{
		NewBaseFrame(path),
	}
}

func (frame *IfElseFrame) GenBeginning(content []byte) []byte {
	return content
}

func (frame *IfElseFrame) GenEnding(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genLineCodeWithStringArg("Collect", frame.path))
	buf.Write(content)
	return buf.Bytes()
}
