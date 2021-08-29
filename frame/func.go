package frame

import (
	"bytes"
)

type FuncFrame struct {
	*baseFrame
	hasResult bool
}

func NewFuncFrame(path string) *FuncFrame {
	return &FuncFrame{baseFrame: NewBaseFrame(path)}
}

// MarkResult mark a function has result
func (frame *FuncFrame) MarkResult() {
	frame.hasResult = true
}

func (frame *FuncFrame) GenBeginning(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(content)
	buf.WriteString(genLineCodeWithStringArg("Collect", "call "+frame.path))
	return buf.Bytes()
}

func (frame *FuncFrame) GenEnding(content []byte) []byte {
	// If a function has a return value, but does not end with return,
	// it means it's impossible to run to here
	if frame.hasResult && !frame.isReturn {
		return content
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString(genLineCodeWithStringArg("Collect", frame.path))
	buf.Write(content)
	return buf.Bytes()
}

func (frame *FuncFrame) String() string {
	str := frame.baseFrame.String()
	if frame.hasResult {
		str += " [result]"
	}
	return str
}
