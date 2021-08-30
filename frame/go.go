package frame

import (
	"bytes"
)

type GoFuncFrame struct {
	*baseFrame
	target string // target function name
}

func NewGoFuncFrame(path string) *GoFuncFrame {
	return &GoFuncFrame{baseFrame: NewBaseFrame(path)}
}

func (frame *GoFuncFrame) SetTarget(target string) {
	frame.target = target
}

func (frame *GoFuncFrame) GenBeginning(content []byte) []byte {
	buf := bytes.NewBuffer(nil)
	if frame.target != "" { // go function(xxx) => go func(){function(xxx)}
		buf.Write(
			bytes.Replace(content, []byte(frame.target),
				[]byte("func(){"+genLineCodeWithStringArg("Collect", "bind "+frame.path)+frame.target),
				1),
		)
		buf.WriteString("}()")
	} else {
		buf.Write(content)
		buf.WriteString(genLineCodeWithStringArg("Collect", "bind "+frame.path))
	}

	return buf.Bytes()
}

func (frame *GoFuncFrame) GenEnding(content []byte) []byte {
	return content
}
