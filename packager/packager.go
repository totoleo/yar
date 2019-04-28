package packager

import (
	"errors"
	"strings"
)

var JSON = [8]byte{'J', 'S', 'O', 'N'}
var MSGP = [8]byte{'M', 'S', 'G', 'P', 'A', 'C', 'K'}

type Packager interface {
	Unmarshal(data []byte, x interface{}) error
	Marshal(interface{}) ([]byte, error)
	GetName() [8]byte
}

type PackFunc func(v interface{}) ([]byte, error)

type UnpackFunc func(data []byte, v interface{}) error

func Pack(name []byte, v interface{}) ([]byte, error) {

	sb := strings.Builder{}
	sb.Write(name)
	s := strings.ToLower(sb.String())

	if strings.Contains(s, "json") {
		return JsonPack(v)
	} else if strings.Contains(s, "msgp") {
		return MsgpPack(v)
	}

	return nil, errors.New("unsupported packager")

}

func Unpack(name []byte, data []byte, v interface{}) error {

	sb := strings.Builder{}
	sb.Write(name)
	s := strings.ToLower(sb.String())

	if strings.Contains(s, "json") {
		return JsonUnpack(data, v)
	} else if strings.Contains(s, "msgp") {
		return MsgpUnpack(data, v)
	}

	return errors.New("unsupported packager")
}
