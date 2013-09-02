package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/dep/packages"
)

var _ = fmt.Print

func getPkg(path string) string {
	p := packages.Get(path)

	b, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

func main() {
	fmt.Println(
		getPkg("github.com/metakeule/dep/example/p"),
		getPkg("github.com/metakeule/dep/example/q"),
		getPkg("fmt"),
	)
}
