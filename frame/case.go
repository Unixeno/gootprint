package frame

import "bytes"

// CaseFrame switch, select
type CaseFrame struct {
	*baseFrame
}

func NewCaseFrame(path string) *CaseFrame {
	return &CaseFrame{NewBaseFrame(path)}
}

func (frame *CaseFrame) GenBeginning(content []byte) []byte {
	return content
}

func (frame *CaseFrame) GenEnding(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genLineCodeWithStringArg("Collect", frame.path))
	buf.Write(content)
	return buf.Bytes()
}