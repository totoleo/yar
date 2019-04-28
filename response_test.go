package yar

import (
	"encoding/json"
	"testing"
)

func TestException_UnmarshalJSON(t *testing.T) {
	data := []byte(`{"i":2596996162,"s":0,"r":"a,b,c,{\"d\":\"leo\",\"0\":\"hello world\",\"1\":\"2019\\\/04\\\/27 15:15:15\"}"}`)
	res := new(Response)
	if err := json.Unmarshal(data, res); err != nil {
		t.Error(err)
	}
}
