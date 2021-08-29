package frame

import "fmt"

type Frame interface {
	HeadBeginning() int                           // line number of the block beginning,
	BodyBeginning() int                           // line number of {
	BodyEnding() int                              // line number of }, or the return statement
	SetReturn(line int)                           // mark the frame has an explicit return
	Path() string                                 // unique frame path
	IsReturn() bool                               // whether this block contains an explicit return statement
	GetInner(index int) Frame                     // get inner frame
	Len() int                                     // amount of inner frames
	Append(Frame)                                 // append an inner frame
	SetPosLine(headBegin, bodyBegin, bodyEnd int) // set line number

	GenBeginning(content []byte) []byte // generator function for BodyBeginning
	GenEnding(content []byte) []byte    // generator function for BodyEnding

	fmt.Stringer
}

func Visit(frame Frame, order int, callback func(Frame)) {
	if order == VisitPreOrder {
		callback(frame)
		for index := 0; index < frame.Len(); index++ {
			Visit(frame.GetInner(index), order, callback)
		}
	} else {
		for index := 0; index < frame.Len(); index++ {
			Visit(frame.GetInner(index), order, callback)
		}
		callback(frame)
	}
}
