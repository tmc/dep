#!/usr/bin/env bash
mkdir -p ~/.dep/tests/gopath/src
mkdir -p ~/.dep/tests/gopath/pkg
mkdir -p ~/.dep/tests/gopath/bin
export GOPATH=~/.dep/tests/gopath
go get github.com/go-dep/dep/depcore
go get github.com/go-dep/deptest_compatible
cd $GOPATH/src/github.com/go-dep/deptest_mod && git checkout 68b5f6
go get github.com/go-dep/deptest_partial_working
go get github.com/go-dep/deptest_partial_failing
go get github.com/go-dep/deptest_incompatible
go get github.com/go-dep/deptest_missing
go get github.com/go-dep/deptest_partial
cd $GOPATH/src/github.com/go-dep/deptest_compatible && go test
cd $GOPATH/src/github.com/go-dep/deptest_incompatible && go test
cd $GOPATH/src/github.com/go-dep/deptest_missing && go test
cd $GOPATH/src/github.com/go-dep/deptest_partial_working && go test
cd $GOPATH/src/github.com/go-dep/deptest_partial_failing && go test
cd $GOPATH/src/github.com/go-dep/deptest_partial && go test
rm -rf ~/.dep/tests/gopath