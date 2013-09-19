package depcore

import (
	"encoding/json"
	"regexp"
)

func toJson(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return b
}

var bzrRevRe = regexp.MustCompile(`revision-id:\s*([^\s]+)`)

// maps a package path to a vcs and a revision
type revision struct {
	VCM    string
	Rev    string
	Parent string
	Tag    string // TODO check if revision is a tag and put it into the rev
}

var revFileName = "dep-rev.json"

func _repoRoot(dir string) string {
	_, root, err := vcsForDir(dir)
	if err != nil {
		panic("can't find repodir for " + dir + " : " + err.Error())
	}
	return root
}
