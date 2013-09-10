package p

const (
	C     = "const"
	I int = iota
)

var VarSval = "huho"
var VarS string
var Z = &TypeStruct{A: "hu"}

var k = "ho"
var Q = "hohu"

var (
	X float32 = 3.4
)

type Int int

type String string

type TypeStruct struct {
	A, B string
	C    int
	D    *TypeStruct
	E    TypeInterface
	F    interface {
		String() string
	}
}

func (ø *TypeStruct) String() string {
	return ø.A
}

type TypeInterface interface {
	Name() (s string)
	Pos() interface{}
	Other(i interface{}) string
	Parent() TypeInterface
	Other2() (string, error)
	In(a string, b int)
	In2(a, b string)
	In3(string, int)
	InterFDefIn(i interface {
		String() string
	})
}

func Func(i interface{}) string {
	return "huho"
}

func ByteFuncIn(b []byte) {
}

func ByteFuncOut() []byte {
	return []byte("hu")
}

type hiddenExport struct{}

func (ø *hiddenExport) Public() {}

func MakeHidden() *hiddenExport {
	return &hiddenExport{}
}

type hidden struct{}

func (ø *hidden) Public() {}
