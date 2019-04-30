package packager

import (
	"bytes"
	"encoding/json"
)

type jsonPack [8]byte

func (p *jsonPack) Unmarshal(data []byte, x interface{}) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	err := d.Decode(x)
	return err
}

func (p *jsonPack) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (p *jsonPack) GetName() [8]byte {
	return *p
}
