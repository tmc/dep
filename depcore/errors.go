package depcore

type ErrorCode int

const (
	_           = iota
	_ ErrorCode = iota
	ErrorGOROOTNotSet
	ErrorGOPATHNotSet
	ErrorDEPTMPNotSet
	ErrorGOPATHInvalid
	ErrorInvalidOptions
	ErrorPackageInternal
	ErrorPackageInvalid
	ErrorPackageNotInGOPATH
	ErrorDirNotAPackage
	ErrorDependancyNotInPackageDir
	ErrorDependancyNotInGOPATH
	ErrorDependancyInfosCorrupt
	ErrorUpdateConflict
)

var ErrorCodeInfos = map[ErrorCode]string{}

var _cleanup = []func(){}

func addCleanup(fn func()) {
	_cleanup = append(_cleanup, fn)
}

func (ø ErrorCode) Error() string {
	return ErrorCodeInfos[ø]
}

// reverse executing all cleanups
func Cleanup() {
	for i := len(_cleanup) - 1; i >= 0; i-- {
		_cleanup[i]()
	}
}
