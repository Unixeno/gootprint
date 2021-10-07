package frame

import "bytes"

// CaseFrame switch, select
type CaseFrame struct {
	*baseFrame
	varName string
}

func NewCaseFrame(path string) *CaseFrame {
	return &CaseFrame{baseFrame: NewBaseFrame(path)}
}

func (frame *CaseFrame) GenBeginning(genEnv *baseEnv, content []byte) []byte {
	return content
}

func (frame *CaseFrame) GenEnding(genEnv *baseEnv, content []byte) []byte {
	frame.varName = genEnv.genPointVarName()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genEnv.genCollect(frame.varName))
	buf.Write(content)
	return buf.Bytes()
}

func (frame *CaseFrame) GenEnv(genEnv *baseEnv) []byte {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(genEnv.genPoint(frame.varName, frame.getStdPath()))
	return buffer.Bytes()
}
