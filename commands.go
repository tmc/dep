package dep

import (
	"flag"
)

func _init(fs *flag.FlagSet) {
	fs.Args()
	dir := fs.Arg(0)
	b, internal := scan(dir)
	if internal {
		panic("can't init internal package " + dir)
	}
	err := writeDepFile(dir, b)
	if err != nil {
		panic(err.Error())
	}
}

func _info(fs *flag.FlagSet) {
	dir := fs.Arg(0)
	b, _ := scan(dir)
	fmt.Printf("%s", b)
}

func _register(fs *flag.FlagSet) {
	dir := fs.Arg(0)
	b, internal := scan(dir)
	err := writeRegisterFile(dir, b, internal)
	if err != nil {
		panic(err.Error())
	}
}

func _diff(fs *flag.FlagSet) {
}

func _update(fs *flag.FlagSet) {
}

func _fix(fs *flag.FlagSet) {
}

func _install(fs *flag.FlagSet) {
}

func init() {
	addCommand("info", _info)
	addCommand("init", _init)
	addCommand("register", _register)
	addCommand("diff", _diff)
	addCommand("update", _update)
	addCommand("fix", _fix)
	addCommand("install", _install)
}
