package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/totoleo/yar"
	"github.com/totoleo/yar/packager"
)

var debug = false

type Client struct {
	hostname string
	net      string

	client *http.Client
	packer packager.Packager
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

	if params == nil {
		r.Params = []interface{}{}
	} else {
		r.Params = params
	}

	r.Method = method

	r.Protocol.MagicNumber = yar.MagicNumber
	r.Protocol.Id = r.Id
	return r, nil
}

func (client *Client) packRequest(r *yar.Request) ([]byte, *yar.Error) {

	r.Protocol.Packager = client.packer.GetName()
	pack, err := client.packer.Marshal(r)

	if err != nil {
		return nil, yar.NewError(yar.ErrorPackager, err.Error())
	}

	return pack, nil
}

func (client *Client) readResponse(reader io.Reader, ret interface{}) *yar.Error {

	allBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return yar.NewError(yar.ErrorResponse, "Read Response Error:"+err.Error())
	}

	if len(allBody) < (yar.ProtocolLength + yar.PackagerLength) {
		return yar.NewError(yar.ErrorResponse, "Response Parse Error:"+string(allBody))
	}

	protocolBuffer := allBody[0 : yar.ProtocolLength+yar.PackagerLength]

	protocol := yar.NewHeader()

	payload := bytes.NewBuffer(protocolBuffer)

	_ = protocol.Read(payload)

	bodyLength := protocol.BodyLength - yar.PackagerLength

	if uint32(len(allBody)-(yar.ProtocolLength+yar.PackagerLength)) < uint32(bodyLength) {
		return yar.NewError(yar.ErrorResponse, "Response Content Error:"+string(allBody))
	}

	bodyBuffer := allBody[yar.ProtocolLength+yar.PackagerLength:]

	response := new(yar.Response)
	err = client.packer.Unmarshal(bodyBuffer, response)

	if err != nil {
		return yar.NewError(yar.ErrorPackager, "Unpack Error:"+err.Error())
	}

	if response.Status != yar.ERR_OKEY {
		return yar.NewError(yar.ErrorResponse, response.Error.GetMessage())
	}

	if ret != nil {

		err = client.packer.Unmarshal(response.Response, ret)

		if err != nil {
			return yar.NewError(yar.ErrorPackager, "pack response ret val error:"+err.Error()+" "+string(allBody))
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

	packBody, err := client.packRequest(r)

	if err != nil {
		return err
	}

	r.Protocol.BodyLength = uint32(len(packBody) + yar.PackagerLength)

	postBuffer := bytes.NewBuffer(nil)

	if err := r.Protocol.Write(postBuffer); err != nil {
		return yar.NewError(yar.ErrorNetwork, err.Error())
	}

	postBuffer.Write(packBody)

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

	responseErr := client.readResponse(resp.Body, ret)
	return responseErr
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
