package frame

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type baseFrame struct {
	headBegin  int     // line number of the block beginning, is the position of first token in the block
	bodyBegin  int     // line number of {
	bodyEnd    int     // line number of }, or the return statement
	path       string  // frame path, is the unique name of a frame
	InnerFrame []Frame // child block in current block
	isReturn   bool    // whether this block contains an explicit return statement
}

func NewBaseFrame(path string) *baseFrame {
	return &baseFrame{
		path:       path,
		InnerFrame: make([]Frame, 0, 8),
	}
}

func (frame *baseFrame) SetPosLine(headBegin, bodyBegin, bodyEnd int) {
	frame.headBegin = headBegin
	frame.bodyBegin = bodyBegin
	frame.bodyEnd = bodyEnd
}

func (frame *baseFrame) Len() int {
	return len(frame.InnerFrame)
}

func (frame *baseFrame) GetInner(index int) Frame {
	return frame.InnerFrame[index] // todo check slice length before access
}

func (frame *baseFrame) Append(inner Frame) {
	frame.InnerFrame = append(frame.InnerFrame, inner)
}

func (frame *baseFrame) HeadBeginning() int {
	return frame.headBegin
}

func (frame *baseFrame) BodyBeginning() int {
	return frame.bodyBegin
}

func (frame *baseFrame) BodyEnding() int {
	return frame.bodyEnd
}

func (frame *baseFrame) SetReturn(line int) {
	frame.bodyEnd = line
	frame.isReturn = true
}

func (frame *baseFrame) Path() string {
	return frame.path
}

func (frame *baseFrame) IsReturn() bool {
	return frame.isReturn
}

func (frame *baseFrame) GenBeginning(content []byte) []byte {
	log.Error("implement me: ", frame.path)
	return content
}

func (frame *baseFrame) GenEnding(content []byte) []byte {
	log.Error("implement me: ", frame.path)
	return content
}

func (frame *baseFrame) String() string {
	str := fmt.Sprintf("%s %d{%d:%d}", frame.path, frame.headBegin, frame.bodyBegin, frame.bodyEnd)
	if frame.isReturn {
		str += " [return]"
	}
	return str
}