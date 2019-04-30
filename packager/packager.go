package packager

var JSON = jsonPack{'J', 'S', 'O', 'N'}
var MSGP = msgpPack{'M', 'S', 'G', 'P', 'A', 'C', 'K'}

type Packager interface {
	Unmarshal(data []byte, x interface{}) error
	Marshal(interface{}) ([]byte, error)
	GetName() [8]byte
}

type PackFunc func(v interface{}) ([]byte, error)

type UnpackFunc func(data []byte, v interface{}) error
