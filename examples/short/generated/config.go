// Code generated by cfgen. DO NOT EDIT.

package generated

import (
	"time"
)

type base struct{}

func (base) APIKey() string             { return "" }
func (base) APIEndpoint() string        { return "" }
func (base) Headers() map[string]string { return (map[string]string)(nil) }
func (base) NowFn() func() time.Time    { return (func() time.Time)(nil) }

func New() Config {
	return base{}
}

type Config interface {
	APIKey() string
	APIEndpoint() string
	Headers() map[string]string
	NowFn() func() time.Time
}

type cfgAPIKey struct {
	Config
	aPIKey string
}

func (cfg cfgAPIKey) APIKey() string {
	return cfg.aPIKey
}

func SetAPIKey(cfg Config, a string) Config {
	return cfgAPIKey{
		Config: cfg,
		aPIKey: a,
	}
}

type cfgAPIEndpoint struct {
	Config
	aPIEndpoint string
}

func (cfg cfgAPIEndpoint) APIEndpoint() string {
	return cfg.aPIEndpoint
}

func SetAPIEndpoint(cfg Config, a string) Config {
	return cfgAPIEndpoint{
		Config:      cfg,
		aPIEndpoint: a,
	}
}

type cfgHeaders struct {
	Config
	headers map[string]string
}

func (cfg cfgHeaders) Headers() map[string]string {
	return cfg.headers
}

func SetHeaders(cfg Config, h map[string]string) Config {
	return cfgHeaders{
		Config:  cfg,
		headers: h,
	}
}

type cfgNowFn struct {
	Config
	nowFn func() time.Time
}

func (cfg cfgNowFn) NowFn() func() time.Time {
	return cfg.nowFn
}

func SetNowFn(cfg Config, n func() time.Time) Config {
	return cfgNowFn{
		Config: cfg,
		nowFn:  n,
	}
}
