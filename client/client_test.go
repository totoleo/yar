package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/totoleo/yar"
	"github.com/totoleo/yar/packager"
)

var __tClient *Client

func TestMain(m *testing.M) {
	tracer := mocktracer.New()
	defer func() {
		//log.Println(tracer.FinishedSpans())
	}()
	opentracing.SetGlobalTracer(tracer)

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		reqHead := yar.NewHeader()
		if err := reqHead.ReadFrom(req.Body); err != nil {
			fmt.Println("err", err)
		}
		yarReq := yar.NewRequest()

		if err := packager.JSON.Unmarshal(req.Body, yarReq); err != nil {
			fmt.Println("unmarshal req err", err)
		}

		var rets = struct {
			Errno   int
			Message string
			Data    struct {
				St   int
				Msg  string
				Data interface{}
			}
		}{
			Errno:   0,
			Message: "success",
			Data: struct {
				St   int
				Msg  string
				Data interface{}
			}{
				St: 0, Msg: "ok", Data: map[string]interface{}{
					"hello": 123, "world": []string{"leo", "coco"}, "params": yarReq.Params,
				},
			},
		}
		x, _ := json.Marshal(rets)
		ress := yar.Response{}
		ress.Status = yar.ERR_OKEY
		ress.Response = x
		out, _ := json.Marshal(ress)

		header := yar.NewHeader()
		header.Id = 1
		header.Packager = packager.JSON
		header.BodyLength = uint32(len(out)) + 8
		if err := header.WriteTo(res); err != nil {
			fmt.Println(err)
		}
		_, err := res.Write(out)
		if err != nil {
			fmt.Println(err)
		}
	}))

	c, err := NewClient(server.URL)

	if err != nil {
		fmt.Println("error", err)
	}
	__tClient = c

	m.Run()
}
func TestCall(t *testing.T) {
	var ret interface{}

	if callErr := __tClient.Call(context.TODO(), "api", &ret, "a", "b", "c", map[string]string{"d": "leo"}); callErr != nil {
		t.Error("error", callErr)
	} else {
		t.Log("Data", ret)
	}
}
func BenchmarkYar(b *testing.B) {
	b.ReportAllocs()

	var ret json.RawMessage
	ctx := context.TODO()
	params := map[string]string{"d": "leo"}
	for i := 0; i < b.N; i++ {

		if callErr := __tClient.Call(ctx, "api", &ret, "a", "b", "c", params); callErr != nil {
			b.Error("error", callErr)
		} else {
			//b.Log("Data", ret)
		}
	}
}
