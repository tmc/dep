package db

type Pkg struct {
	Package    string
	JsonMd5    string
	Json       []byte
	ImportsMd5 string
	ExportsMd5 string
	InitMd5    string
}

type Exp struct {
	Package string
	Name    string
	Value   string
}

type Imp struct {
	Package string
	Import  string
	Name    string
	Value   string
}
