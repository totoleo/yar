package packager

import (
	"gopkg.in/vmihailenco/msgpack.v2"
)

type msgpPack [8]byte

func (p *msgpPack) Unmarshal(data []byte, x interface{}) error {
	return msgpack.Unmarshal(data, x)
}
func (p *msgpPack) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}
func (p *msgpPack) GetName() [8]byte {
	return *p
}
