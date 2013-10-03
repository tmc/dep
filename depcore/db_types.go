package depcore

type dbPkg struct {
	Package string
	JsonMd5 string
	Json    []byte
	InitMd5 string
}

type exp struct {
	Package string
	Name    string
	Value   string
}

type imp struct {
	Package string
	Import  string
	Name    string
	Value   string
}
