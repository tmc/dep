package main

import (
	"github.com/codegangsta/cli"
)

// func(c *cli.Context) ErrorCode
// (c *cli.Context) ErrorCode
func _init(c *cli.Context) ErrorCode {
	//fs.Args()
	//dir := fs.Arg(0)
	/*
		b, internal := scan(dir)
		if internal {
			panic("can't init internal package " + dir)
		}
		err := writeDepFile(dir, b)
		if err != nil {
			panic(err.Error())
		}
	*/
	return 0
}

func _info(c *cli.Context) ErrorCode {
	/*
		dir := fs.Arg(0)
		b, _ := scan(dir)
		fmt.Printf("%s", b)
	*/
	return 0
}

func _register(c *cli.Context) ErrorCode {
	/*
		dir := fs.Arg(0)
		b, internal := scan(dir)
		err := writeRegisterFile(dir, b, internal)
		if err != nil {
			panic(err.Error())
		}
	*/
	return 0
}

func _diff(c *cli.Context) ErrorCode {
	return 0
}

func _store(c *cli.Context) ErrorCode {
	return 0
}

func _update(c *cli.Context) ErrorCode {
	return 0
}

func _fix(c *cli.Context) ErrorCode {
	return 0
}

func _install(c *cli.Context) ErrorCode {
	return 0
}

func _get(c *cli.Context) ErrorCode {
	return 0
}

func _lint(c *cli.Context) ErrorCode {
	return 0
}
