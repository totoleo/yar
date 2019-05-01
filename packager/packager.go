package packager

import (
	"bytes"
	"io"
)

var JSON = jsonPack{'J', 'S', 'O', 'N'}
var MSGP = msgpPack{'M', 'S', 'G', 'P', 'A', 'C', 'K'}

type Packager interface {
	Unmarshal(reader io.Reader, x interface{}) error
	Marshal(interface{}) (*bytes.Buffer, error)
	GetName() [8]byte
}

type PackFunc func(v interface{}) ([]byte, error)

type UnpackFunc func(data []byte, v interface{}) error
