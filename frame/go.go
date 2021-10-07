package frame

import (
	"bytes"
	"fmt"
)

type GoFuncFrame struct {
	*baseFrame
	target    string // target function name, empty means anonymous function
	callEvent string
	eventVar  string
}

func NewGoFuncFrame(path string) *GoFuncFrame {
	return &GoFuncFrame{baseFrame: NewBaseFrame(path)}
}

func (frame *GoFuncFrame) SetTarget(target string) {
	frame.target = target
}

func (frame *GoFuncFrame) GenBeginning(genEnv *baseEnv, content []byte) []byte {
	genEnv.NewFuncEnv()
	buf := bytes.NewBuffer(nil)
	if frame.target != "" { // go function(xxx) => go func(){function(xxx)}
		replaceTarget := fmt.Sprintf("func(){%s%s", genEnv.genBind(), frame.target)
		buf.Write(
			bytes.Replace(content, []byte(frame.target),
				[]byte(replaceTarget),
				1),
		)
		buf.WriteString("}()")
	} else { // empty means target is an anonymous function, treat as a normal function
		buf.Write(content)
		buf.WriteString(genEnv.genBind())
		frame.callEvent = genEnv.genPointVarName()
		buf.WriteString(genEnv.genCall(genEnv.GetCurrentGoIDVarName(), frame.callEvent))
	}

	return buf.Bytes()
}

func (frame *GoFuncFrame) GenEnding(genEnv *baseEnv, content []byte) []byte {
	defer genEnv.PopFuncEnv()
	// we don't need to generate trace code for go func call if target function isn't a Func Lit
	if frame.target != "" {
		return content
	}
	// as go called function must have no return value, we can add trace code to all anonymous function
	buf := bytes.NewBuffer(nil)
	frame.eventVar = genEnv.genPointVarName()
	buf.WriteString(genEnv.genCollect(frame.eventVar))
	buf.Write(content)
	return buf.Bytes()
}

func (frame *GoFuncFrame) GenEnv(genEnv *baseEnv) []byte {
	buffer := bytes.NewBuffer(nil)
	if frame.callEvent != "" {
		buffer.WriteString(genEnv.genPoint(frame.callEvent, frame.getStdPath()))
	}
	if frame.eventVar != "" {
		buffer.WriteString(genEnv.genPoint(frame.eventVar, frame.getStdPath()))
	}
	return buffer.Bytes()
}
