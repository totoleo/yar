package packager

import (
	"bytes"
	"encoding/json"
	"io"
)

type jsonPack [8]byte

func (p *jsonPack) Unmarshal(reader io.Reader, x interface{}) error {
	d := json.NewDecoder(reader)
	d.UseNumber()
	return d.Decode(x)
}

func (p *jsonPack) Marshal(v interface{}) (*bytes.Buffer, error) {
	w := bytes.NewBuffer(make([]byte, 0, 32))
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return w, encoder.Encode(v)
}

func (p *jsonPack) GetName() [8]byte {
	return *p
}
