package main

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/totoleo/yar/packager"
	"testing"

	"github.com/totoleo/yar/client"
)

var tClient *client.Client

func TestMain(m *testing.M) {
	tracer := mocktracer.New()
	defer func() {
	}()
	opentracing.SetGlobalTracer(tracer)
	c, err := client.NewClient("http://cap.dev:8001/app")

	if err != nil {
		fmt.Println("error", err)
	}
	//这是默认值
	c.Opt.Packager = packager.JSON
	//这是默认值
	c.Opt.Encrypt = false

	tClient = c

	m.Run()
}
func TestCall(t *testing.T) {
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
