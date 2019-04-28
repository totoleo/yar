package packager

import (
	"bytes"
	"encoding/json"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type msgpPack [8]byte

func (p *msgpPack) Unmarshal(data []byte, x interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	err := d.Decode(x)
	return err
}
func (p *msgpPack) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
func (p *msgpPack) GetName() [8]byte {
	return *p
}

func MsgpPack(v interface{}) ([]byte, error) {

	return msgpack.Marshal(v)
}

func MsgpUnpack(data []byte, v interface{}) error {

	return msgpack.Unmarshal(data, v)

}
