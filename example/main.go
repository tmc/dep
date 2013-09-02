package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/dep/packages"
	"github.com/metakeule/exports"
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
		//		getPkg("github.com/metakeule/dep/example/p"),
		getPkg("github.com/metakeule/dep/example/q"),
		//		getPkg("fmt"),
	)

	_ = exports.SelectorExpressions

	/*
		expr := exports.SelectorExpressions("github.com/metakeule/dep/example/q")

		for k, _ := range expr {
			fmt.Printf("%s.%s\n", k[0], k[1])
		}
	*/
	// fmt.Printf("%v", expr)
}
