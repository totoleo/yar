package main

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
	"github.com/totoleo/yar/client"
	"github.com/totoleo/yar/packager"
)

var tClient *client.Client

func TestMain(m *testing.M) {
	tracer := mocktracer.New()
	defer func() {
	}()
	opentracing.SetGlobalTracer(tracer)
	var rets = []string{
		"hello", "world",
	}
	out, _ := json.Marshal(rets)
	ress := yar.Response{}
	ress.Status = yar.ERR_OKEY
	ress.Response = out
	out, _ = json.Marshal(ress)

	header := yar.NewHeader()
	header.Id = 1
	header.Packager = packager.JSON
	header.BodyLength = uint32(len(out))

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		_ = header.Write(res)
		_, _ = res.Write(out)
	}))

	c, err := client.NewClient(server.URL)

	if err != nil {
		fmt.Println("error", err)
	}
	tClient = c

	m.Run()
}
func TestCall(t *testing.T) {
	t.Helper()
	var ret interface{}

	if callErr := tClient.Call(context.TODO(), "api", &ret, "a", "b", "c", map[string]string{"d": "leo"}); callErr != nil {
		t.Error("error", callErr)
	} else {
		t.Log("data", ret)
	}
}
func BenchmarkYar(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var ret interface{}

		if callErr := tClient.Call(context.TODO(), "api", &ret, "a", "b", "c", map[string]string{"d": "leo"}); callErr != nil {
			b.Error("error", callErr)
		} else {
			//b.Log("data", ret)
		}
	}
}
