package frame

import (
	"bytes"
)

type ForFrame struct {
	*baseFrame
}

func NewForFrame(path string) *ForFrame {
	return &ForFrame{NewBaseFrame(path)}
}


func (frame *ForFrame) GenBeginning(content []byte) []byte {
	return content
}

func (frame *ForFrame) GenEnding(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genLineCodeWithStringArg("Collect", frame.path))
	buf.Write(content)
	return buf.Bytes()
}