package frame

import (
	"bytes"
)

type GoFuncFrame struct {
	*baseFrame
	target string // target function name, empty means anonymous function
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
	} else { // empty means target is an anonymous function, treat as a normal function
		buf.Write(content)
		buf.WriteString(genLineCodeWithStringArg("Collect", "bind "+frame.path))
		buf.WriteString(genLineCodeWithStringArg("Collect", "call "+frame.path))
	}

	return buf.Bytes()
}

func (frame *GoFuncFrame) GenEnding(content []byte) []byte {
	// we don't need to generate trace code for go func call if target function isn't a Func Lit
	if frame.target != "" {
		return content
	}
	// as go called function must have no return value, we can add trace code to all anonymous function
	buf := bytes.NewBuffer(nil)
	buf.WriteString(genLineCodeWithStringArg("Collect", frame.path))
	buf.Write(content)
	return buf.Bytes()
}
