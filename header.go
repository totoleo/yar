package yar

import (
	"encoding/binary"
	"io"
	"sync"
)

const (
	ProtocolLength = 90
	PackagerLength = 8
)

type ErrorType int

const (
	MagicNumber = 0x80DFEC60
)

var YarProvider = [28]byte{'Y', 'a', 'r', '/', 'G', 'o', ' ', 'v', '1', '.', '0', '.', '0'}

const (
	ERR_OKEY           ErrorType = 0x0
	ERR_PACKAGER       ErrorType = 0x1
	ERR_PROTOCOL       ErrorType = 0x2
	ERR_REQUEST        ErrorType = 0x4
	ERR_OUTPUT         ErrorType = 0x8
	ERR_TRANSPORT      ErrorType = 0x10
	ERR_FORBIDDEN      ErrorType = 0x20
	ERR_EXCEPTION      ErrorType = 0x40
	ERR_EMPTY_RESPONSE ErrorType = 0x80
)

type Header struct {
	Id          uint32   //4
	Version     uint16   //2
	MagicNumber uint32   //4
	Reserved    uint32   //4
	Provider    [28]byte //28
	Encrypt     uint32   //4
	Token       [32]byte //32
	BodyLength  uint32   //4
	Packager    [8]byte  //8
}

var headerPool = sync.Pool{
	New: func() interface{} {
		return NewHeader()
	},
}

func GetHeader() *Header {
	return headerPool.Get().(*Header)
}
func NewHeader() *Header {
	proto := new(Header)
	proto.MagicNumber = MagicNumber
	proto.Provider = YarProvider
	return proto
}
func Return(h *Header) {
	headerPool.Put(h)
}

func (h *Header) ReadFrom(payload io.Reader) error {
	return binary.Read(payload, binary.BigEndian, h)
}

func (h *Header) WriteTo(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, h)
}
