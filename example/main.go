package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/exports"
	// "github.com/metakeule/exports"
)

var _ = fmt.Print

var Env = exports.DefaultEnv

func getPkg(path string) string {
	p := Env.Pkg(path)

	b, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

//var pkg = "github.com/metakeule/dep/example/p"
var pkg = "github.com/metakeule/dep/example/q"

func main() {
	//fmt.Printf("%v\n", exports.GetUsedImports(pkg))
	fmt.Println(getPkg(pkg))
	/*
		fmt.Println(
			//		getPkg("github.com/metakeule/dep/example/p"),
			getPkg("github.com/metakeule/dep/example/q"),
			//		getPkg("fmt"),
		)
	*/
	//	_ = exports.SelectorExpressions

	/*
		expr := exports.SelectorExpressions("github.com/metakeule/dep/example/q")

		for k, _ := range expr {
			fmt.Printf("%s.%s\n", k[0], k[1])
		}
	*/
	// fmt.Printf("%v", expr)
}
