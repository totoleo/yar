package yar

import "github.com/totoleo/yar/packager"

type YarOpt int

const (
	LogLevelDebug  int = 0x0001
	LoglevelNormal int = 0x0002
	LogLevelError  int = 0x0004
)

type Opt struct {
	MagicNumber uint32
	Packager    [8]byte
	Encrypt     bool
	LogLevel    int
}

func NewOpt() *Opt {
	opt := new(Opt)
	opt.MagicNumber = MagicNumber
	opt.Encrypt = false
	opt.Packager = packager.JSON
	opt.LogLevel = LogLevelError
	return opt
}
