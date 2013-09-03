package q

import (
	`fmt`
	"github.com/metakeule/dep/example/p"
	i "io"
)

var A = p.VarSval
var B = fmt.Print

type C struct{}

var X = &p.TypeStruct{}

func P(i.Reader) {

}

func Q(*p.TypeStruct) {

}

func u(i.ByteWriter) (shown int) {
	return 0
}

var CC = &C{}
