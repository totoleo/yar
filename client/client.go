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
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"time"
)

var debug = false

type Client struct {
	hostname string
	net      string
	Opt      *yar.Opt

	client *http.Client
}

type Option struct {
	f func(opts *Options)
}

func WithHttpClient(client *http.Client) Option {
	return Option{f: func(opts *Options) {
		opts.Client = client
	}}
}

func WithTimeout(timeout time.Duration) Option {
	return Option{f: func(opts *Options) {
		opts.Timeout = timeout
	}}
}
func WithPackager(pack packager.Packager) Option {
	return Option{f: func(opts *Options) {
		opts.Packer = pack
	}}
}

type Options struct {
	Timeout        time.Duration
	ConnectTimeout uint32
	Packager       string
	LogLevel       int
	Client         *http.Client
	Packer         packager.Packager
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
	client.Opt = yar.NewOpt()

	if opts.Client == nil {

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		tr.DisableKeepAlives = true
		//if Client.Opt.DNSCache == true {
		//	tr.DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
		//		separator := strings.LastIndex(address, ":")
		//		ips, err := globalResolver.Lookup(address[:separator])
		//		if err != nil {
		//			return nil, errors.New("Lookup Error:" + err.Error())
		//		}
		//		if len(ips) < 1 {
		//			return nil, errors.New("lookup Error: No IP Resolver Result Found")
		//		}
		//		var dial net.Dialer
		//		return dial.DialContext(ctx, "tcp", ips[0].String()+address[separator:])
		//	}
		//}
		httpClient := &http.Client{
			Transport: tr,
			Timeout:   1000 * time.Millisecond,
		}
		client.client = httpClient
	}
	{
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

	packagerName := client.Opt.Packager

	r.Protocol.Packager = packagerName
	pack, err := packager.Pack(packagerName[:], r)

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

	protocol.Read(payload)

	bodyLength := protocol.BodyLength - yar.PackagerLength

	if uint32(len(allBody)-(yar.ProtocolLength+yar.PackagerLength)) < uint32(bodyLength) {
		return yar.NewError(yar.ErrorResponse, "Response Content Error:"+string(allBody))
	}

	bodyBuffer := allBody[yar.ProtocolLength+yar.PackagerLength:]

	response := new(yar.Response)
	err = packager.Unpack(client.Opt.Packager[:], bodyBuffer, response)

	if err != nil {
		return yar.NewError(yar.ErrorPackager, "Unpack Error:"+err.Error())
	}

	if response.Status != yar.ERR_OKEY {
		return yar.NewError(yar.ErrorResponse, response.Error.GetMessage())
	}

	if ret != nil {

		err = packager.Unpack(client.Opt.Packager[:], response.Response, ret)

		if err != nil {
			return yar.NewError(yar.ErrorPackager, "pack response ret val error:"+err.Error()+" "+string(allBody))
		}
	}

	return nil
}

func (client *Client) httpHandler(ctx context.Context, method string, ret interface{}, params ...interface{}) *yar.Error {

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
	r.Protocol.Write(postBuffer)
	postBuffer.Write(packBody)

	//todo 停止验证HTTPS请求
	httpClient := client.getClient()

	clientSpan, ctx := opentracing.StartSpanFromContext(ctx, method)

	//uncomment to enable
	if debug {
		trace := NewClientTrace(clientSpan)
		ctx = httptrace.WithClientTrace(ctx, trace)
	}

	req, postErr := http.NewRequest("POST", client.hostname, postBuffer)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	// Set some tags on the clientSpan to annotate that it's the Client span. The additional HTTP tags are useful for debugging purposes.
	ext.SpanKindRPCClient.Set(clientSpan)

	// Inject the Client span context into the headers
	clientSpan.Tracer().Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	resp, postErr := httpClient.Do(req)
	clientSpan.Finish()

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
