package client

import (
	"net/http"
	"time"

	"github.com/totoleo/yar/packager"
)

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
	LogLevel       int
	Client         *http.Client
	Packer         packager.Packager
}
