package requests

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	"io"
	"net/http"
	"time"
)

var logger = utils.GetModuleLogger("request")

type option struct {
	Timeout            time.Duration
	InsecureSkipVerify bool

	Debug               bool
	Cookies             []*http.Cookie
	Header              gout.H
	Proxy               string
	HttpCode            *int
	Retry               int
	ProxyCallbackOption func(out interface{}, proxy string)
}

func (o *option) getGout() *gout.Client {
	var timeoutOpt = gout.WithTimeout(time.Second * 5)
	if o.Timeout != 0 {
		timeoutOpt = gout.WithTimeout(o.Timeout)
	}
	if o.InsecureSkipVerify {
		return gout.NewWithOpt(timeoutOpt, gout.WithInsecureSkipVerify())
	}
	return gout.NewWithOpt(timeoutOpt)
}

type Option func(o *option)

func empty(*option) {}

func HttpCookieOption(cookie *http.Cookie) Option {
	return func(o *option) {
		o.Cookies = append(o.Cookies, cookie)
	}
}

func CookieOption(name, value string) Option {
	return HttpCookieOption(&http.Cookie{Name: name, Value: value})
}

func TimeoutOption(d time.Duration) Option {
	return func(o *option) {
		o.Timeout = d
	}
}

func HttpCodeOption(code *int) Option {
	return func(o *option) {
		o.HttpCode = code
	}
}

func HeaderOption(key, value string) Option {
	return func(o *option) {
		if o.Header == nil {
			o.Header = make(gout.H)
		}
		o.Header[key] = value
	}
}

func AddUAOption() Option {
	return HeaderOption("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
}

func ProxyOption(prefer proxy_pool.Prefer) Option {
	if prefer == proxy_pool.PreferNone {
		return empty
	}
	proxy, err := proxy_pool.Get(prefer)
	if err != nil {
		if err != proxy_pool.ErrNil {
			logger.Errorf("get proxy failed")
		}
		return empty
	} else {
		return func(o *option) {
			o.Proxy = proxy.ProxyString()
		}
	}
}

func RetryOption(retry int) Option {
	return func(o *option) {
		o.Retry = retry
	}
}

func ProxyCallbackOption(f func(out interface{}, proxy string)) Option {
	return func(o *option) {
		o.ProxyCallbackOption = f
	}
}

func DisableTlsOption() Option {
	return func(o *option) {
		o.InsecureSkipVerify = true
	}
}

func DebugOption() Option {
	return func(o *option) {
		o.Debug = true
	}
}

func Do(f func(*gout.Client) *dataflow.DataFlow, out interface{}, options ...Option) error {
	var opt = new(option)
	for _, o := range options {
		o(opt)
	}
	if opt.ProxyCallbackOption != nil && len(opt.Proxy) > 0 {
		defer func() {
			opt.ProxyCallbackOption(out, opt.Proxy)
		}()
	}
	var df = f(opt.getGout())
	if opt.Debug {
		df.Debug(true)
	}
	if len(opt.Cookies) > 0 {
		df.SetCookies(opt.Cookies...)
	}
	if len(opt.Header) > 0 {
		df.SetHeader(opt.Header)
	}
	if len(opt.Proxy) > 0 {
		df.SetProxy(opt.Proxy)
	}
	if opt.HttpCode != nil {
		df.Code(opt.HttpCode)
	}
	switch out.(type) {
	case io.Writer, []byte, *string:
		df.BindBody(out)
	default:
		df.BindJSON(out)
	}
	if opt.Retry > 0 {
		return df.F().Retry().Attempt(opt.Retry).Do()
	}
	return df.Do()
}

func Get(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.GET(url).SetQuery(params)
	}, out, options...)
}

func Post(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.POST(url).SetForm(params)
	}, out, options...)
}

func PostJson(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.POST(url).SetJSON(params)
	}, out, options...)
}
