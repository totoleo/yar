package yar

import (
	"encoding/json"
	"errors"
)

type exception struct {
	Message string `json:"message" msgpack:"message"`
	Code    int    `json:"code" msgpack:"code"`
	File    string `json:"file" msgpack:"file"`
	Line    int    `json:"line" msgpack:"line"`
	Type    string `json:"_type" msgpack:"_type"`
}
type Exception struct {
	e exception
}

func NewException(message string) *Exception {
	return &Exception{e: exception{Message: message}}
}

func (e *Exception) GetMessage() string {
	return e.e.Message
}
func (e *Exception) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if e == nil {
		*e = Exception{e: exception{}}
	}
	if data[0] == '"' {
		e.e.Message = string(data)
		return nil
	} else {
		return json.Unmarshal(data, &e.e)
	}
}

type RawMessage json.RawMessage

// MarshalMsgpack returns *m as the msgpack encoding of m.
func (m *RawMessage) MarshalMsgpack() ([]byte, error) {
	return *m, nil
}

// UnmarshalMsgpack sets *m to a copy of data.
func (m *RawMessage) UnmarshalMsgpack(data []byte) error {
	if m == nil {
		return errors.New("msgpack.RawMessage: UnmarshalMsgpack on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type Response struct {
	Protocol *Header         `json:"_" msgpack:"_"`
	Id       uint32          `json:"i" msgpack:"i"`
	Error    *Exception      `json:"e" msgpack:"e"`
	Out      json.RawMessage `json:"o" msgpack:"o"`
	Response json.RawMessage `json:"r" msgpack:"r"`
	Status   ErrorType       `json:"s" msgpack:"s"`
}

func NewResponse() (response *Response) {

	response = new(Response)

	return response
}

func (r *Response) Exception(msg string) {

	r.Status = ERR_OUTPUT
	r.Error = NewException(msg)
}
