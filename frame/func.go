package frame

import (
	"bytes"
)

type FuncFrame struct {
	*baseFrame
	hasResult bool
	callEvent string
	goIDEvent string
	eventVar  string
}

func NewFuncFrame(path string) *FuncFrame {
	return &FuncFrame{baseFrame: NewBaseFrame(path)}
}

// MarkResult mark a function has result
func (frame *FuncFrame) MarkResult() {
	frame.hasResult = true
}

func (frame *FuncFrame) GenBeginning(genEnv *baseEnv, content []byte) []byte {
	genEnv.NewFuncEnv()
	frame.callEvent = genEnv.genPointVarName()
	buf := bytes.NewBuffer(nil)
	buf.Write(content)
	buf.WriteString(genEnv.genCall(genEnv.GetCurrentGoIDVarName(), frame.callEvent))
	return buf.Bytes()
}

func (frame *FuncFrame) GenEnding(genEnv *baseEnv, content []byte) []byte {
	defer genEnv.PopFuncEnv()
	// If a function has a return value, but does not end with return,
	// it means it's impossible to run to here
	if frame.hasResult && !frame.isReturn {
		return content
	}

	frame.eventVar = genEnv.genPointVarName()

	buf := bytes.NewBuffer(nil)
	buf.WriteString(genEnv.genCollect(frame.eventVar))
	buf.Write(content)
	return buf.Bytes()
}

func (frame *FuncFrame) GenEnv(genEnv *baseEnv) []byte {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(genEnv.genPoint(frame.callEvent, frame.getStdPath()))
	if frame.eventVar != "" {
		buffer.WriteString(genEnv.genPoint(frame.eventVar, frame.getStdPath()))
	}
	return buffer.Bytes()
}

func (frame *FuncFrame) String() string {
	str := frame.baseFrame.String()
	if frame.hasResult {
		str += " [result]"
	}
	return str
}
