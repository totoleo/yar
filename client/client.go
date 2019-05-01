package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/totoleo/yar"
	"github.com/totoleo/yar/packager"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
)

var debug = false

type Client struct {
	hostname string
	net      string

	client *http.Client
	packer packager.Packager

	globalReqId uint32
}

// 获取一个YAR 客户端
// addr为带请求协议的地址。支持以下格式
// http://xxxxxxxx
// https://xxxx.xx.xx
// tcp://xxxx
// udp://xxxx
func NewClient(addr string, options ...Option) (*Client, *yar.Error) {
	netName, err := parseAddrNetName(addr)
	if err != nil {
		return nil, yar.NewError(yar.ErrorParam, err.Error())
	}
	var opts = new(Options)
	for _, do := range options {
		do.f(opts)
	}

	client := new(Client)

	client.hostname = addr
	client.net = netName

	if opts.Client == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		tr.DisableKeepAlives = true
		httpClient := &http.Client{
			Transport: tr,
		}
		client.client = httpClient
	}
	if opts.Timeout > 0 {
		client.client.Timeout = opts.Timeout
	}
	if opts.Packer != nil {
		client.packer = opts.Packer
	} else {
		client.packer = &packager.JSON
	}
	return client, nil
}

func (client *Client) Call(ctx context.Context, method string, ret interface{}, params ...interface{}) *yar.Error {

	if client.net == "http" || client.net == "https" {
		return client.httpHandler(ctx, method, ret, params...)
	}

	return yar.NewError(yar.ErrorConfig, "unsupported non http protocol")

}

func (client *Client) initRequest(method string, params ...interface{}) (*yar.Request, *yar.Error) {

	r := yar.NewRequest()

	if len(method) < 1 {
		return nil, yar.NewError(yar.ErrorParam, "call empty method")
	}

	r.Id = atomic.AddUint32(&client.globalReqId, 1)
	r.Protocol.Id = r.Id

	if params == nil {
		r.Params = []interface{}{}
	} else {
		r.Params = params
	}

	r.Method = method

	return r, nil
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return yar.NewResponse()
	},
}

func (client *Client) readResponse(reader io.Reader, ret interface{}) *yar.Error {

	protocol := yar.GetHeader()

	if err := protocol.ReadFrom(reader); err != nil {
		return yar.NewError(yar.ErrorResponse, err.Error())
	}

	bodyLength := protocol.BodyLength - yar.PackagerLength
	yar.Return(protocol)

	response := responsePool.Get().(*yar.Response)
	defer responsePool.Put(response)

	err := client.packer.Unmarshal(io.LimitReader(reader, int64(bodyLength)), response)

	if err != nil {
		return yar.NewError(yar.ErrorPackager, "Unpack Error:"+err.Error())
	}

	if response.Status != yar.ERR_OKEY {
		return yar.NewError(yar.ErrorResponse, response.Error.GetMessage())
	}

	if ret != nil {
		err = client.packer.Unmarshal(bytes.NewReader(response.Response), ret)
		if err != nil {
			return yar.NewError(yar.ErrorPackager, "pack response ret val error:"+err.Error())
		}
	}

	return nil
}

func (client *Client) httpHandler(ctx context.Context, method string, ret interface{}, params ...interface{}) *yar.Error {

	clientSpan, ctx := opentracing.StartSpanFromContext(ctx, method)
	defer clientSpan.Finish()

	r, err := client.initRequest(method, params...)

	if err != nil {
		return err
	}

	r.Protocol.Packager = client.packer.GetName()

	w, pacErr := client.packer.Marshal(r)

	if pacErr != nil {
		return yar.NewError(yar.ErrorPackager, pacErr.Error())
	}

	r.Protocol.BodyLength = uint32(w.Len() + yar.PackagerLength)

	postBuffer := bytes.NewBuffer(make([]byte, 0, yar.ProtocolLength))
	if err := r.Protocol.WriteTo(postBuffer); err != nil {
		return yar.NewError(yar.ErrorNetwork, err.Error())
	}

	if _, err := w.WriteTo(postBuffer); err != nil {
		return yar.NewError(yar.ErrorPackager, err.Error())
	}

	httpClient := client.getClient()

	//uncomment to enable
	if debug {
		trace := NewClientTrace(clientSpan)
		ctx = httptrace.WithClientTrace(ctx, trace)
	}

	req, postErr := http.NewRequest("POST", client.hostname, postBuffer)
	if postErr != nil {
		return yar.NewError(yar.ErrorRequest, postErr.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	// Set some tags on the clientSpan to annotate that it's the Client span. The additional HTTP tags are useful for debugging purposes.
	ext.SpanKindRPCClient.Set(clientSpan)

	// Inject the Client span context into the headers
	_ = clientSpan.Tracer().Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	resp, postErr := httpClient.Do(req)

	if postErr != nil {
		return yar.NewError(yar.ErrorNetwork, postErr.Error())
	}

	return client.readResponse(resp.Body, ret)
}

func (client *Client) getClient() *http.Client {
	return client.client
}

func (client *Client) sockHandler(method string, ret interface{}, params ...interface{}) *yar.Error {
	return yar.NewError(yar.ErrorParam, "unsupported sock request")
}

func NewClientTrace(span opentracing.Span) *httptrace.ClientTrace {
	trace := &clientTrace{span: span}
	return &httptrace.ClientTrace{
		WroteHeaderField:     trace.writeHeaderField,
		PutIdleConn:          trace.putIdleConn,
		GotFirstResponseByte: trace.gotFirstResponseByte,
	}
}

// clientTrace holds a reference to the Span and
// provides methods used as ClientTrace callbacks
type clientTrace struct {
	span opentracing.Span
}

func (h *clientTrace) writeHeaderField(key string, vals []string) {
	//if strings.HasPrefix(key, "Mockpfx-") {
	//	return
	//}
	h.span.LogFields(log.Object(key, vals))
}
func (h *clientTrace) putIdleConn(err error) {
	if err != nil {
		h.span.LogFields(log.Error(err))
	}
	h.span.LogFields(log.Bool("putIdleConn", err == nil))
}
func (h *clientTrace) gotFirstResponseByte() {
	h.span.LogFields(log.Object("TTFB", 1))
}
