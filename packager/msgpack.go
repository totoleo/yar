package packager

import (
	"bytes"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
)

type msgpPack [8]byte

func (p *msgpPack) Unmarshal(reader io.Reader, x interface{}) error {
	return msgpack.NewDecoder(reader).Decode(x)
}
func (p *msgpPack) Marshal(v interface{}) (*bytes.Buffer, error) {
	w := bytes.NewBuffer(make([]byte, 0, 32))
	encoder := msgpack.NewEncoder(w)
	return w, encoder.Encode(v)
}
func (p *msgpPack) GetName() [8]byte {
	return *p
}
