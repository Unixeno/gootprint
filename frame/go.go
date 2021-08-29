package frame

type GoFuncFrame struct {
	*baseFrame
}

func NewGoFuncFrame(path string) *GoFuncFrame {
	return &GoFuncFrame{NewBaseFrame(path)}
}
