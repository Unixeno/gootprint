package frame

type Stats struct {
	IfAmount       int
	ForAmount      int
	CaseAmount     int // include select, switch and typed switch
	FuncAmount     int
	GoFuncAmount   int
	InjectionPoint int // number of point to inject track code
	Lines          int // number of lines of code
}

func (s *Stats) Add(b Stats) {
	s.IfAmount += b.IfAmount
	s.ForAmount += b.ForAmount
	s.CaseAmount += b.CaseAmount
	s.FuncAmount += b.FuncAmount
	s.GoFuncAmount += b.GoFuncAmount
	s.InjectionPoint += b.InjectionPoint
	s.Lines += b.Lines
}
