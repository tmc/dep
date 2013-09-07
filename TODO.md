# Concept

## deb Tool

### shared functions:
#### makeMd5(string) string
    - make checksum for string
#### loadJson(string) package
    - loads package info from json and returns a package
#### packageDiff(package a, package b) (string)
    - compage package a and b and return a unified diff
#### getImportRevisions(package string)
    - returns the last working imported revisions
#### getJson(package string)string
    - get config and return it as json
#### writeLocal(package string)
    - get config from getJson
    - get working revisions 
    - get md5 from makeMd5
    - and write dep.json and dep.md5, dep.imports to project directory
#### writeRegistry
    - get config from getJson
    - get md5 from makeMd5
    - and write dep.json and dep.md5, dep.imports to path.Join(goPATH, "dep")

### modes:

#### packgage: do something with a single package
##### init
##### info returns config and working imports
##### register
##### diff
##### update
##### fix: try to find a common revision that all dependant packages may work with for all not working dependancies
##### get: get a package be trying to install it first into the tentative GOPATH
and then check the dependancies, if a revision is known checkout the revision and do it recursively for all package dependancies.

#### all: do something with all packages in GOPATH
##### init
##### info
##### register
##### diff
##### list (list all packages with registered/not-registered status)
##### update
##### fix



#### init
    - writeLocal
    - writeRegistry
#### register
    - writeRegistry
#### info
    - getJson and print it
#### register-all
    - writeRegistry for all files in GOPATH
#### diff
    - make a diff between the registered diff or the dep.json if is not registered and the current
#### not-registered
    - list all packages in GOPATH that are not registered yet
#### registered
    - list all packages in GOPATH that are registered
#### registry-update
    - update all packages in GOPATH in the registry
#### update [package name]
    - get current package to helper gopath in .dep and compare the config
#### diff-all
    make a diff for all packages

## exports:
    - make checksums for init and main functions and export them
    - get the revision for the packages that the current package imports and the branch from that it comes
    - checkout a given revision and a given branch