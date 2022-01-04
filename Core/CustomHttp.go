package Core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CustomClient struct {
	CuClient *http.Client
	Method   string
	Mod      string
	Headers  []HTTPHeader
}

type HTTPHeader struct {
	Name  string
	Value string
}

type Additional struct {
	Mod   string
	Value string
}

func (custom *CustomClient) NewHttpClient(Opt *Options) (*CustomClient, error) {
	custom.CuClient = &http.Client{
		Transport: Opt.Transport,
		Timeout:   time.Second * time.Duration(Opt.Timeout),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	custom.Method = Opt.Method
	custom.Headers = Opt.Headers
	custom.Mod = Opt.Mod

	return custom, nil
}

func (custom *CustomClient) RunRequest(ctx context.Context, Url string, Para Additional) (*ReqRes, error) {
	//defer utils.TimeCost()()

	response, err := custom.DoRequest(ctx, Url, Para)

	result := ReqRes{}

	if err != nil {
		// ignore context canceled errors
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, err
		}
		// 输出错误,暂时看下来context deadline的页面都是无意义的页面,如果以后出现再解决
		//FileLogger.Error("RunRequestErr", zap.String("Error", err.Error()))
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	result.StatusCode = response.StatusCode
	result.Header = response.Header
	result.Length = int64(len(body))
	result.Body = body

	return &result, nil

}

func (custom *CustomClient) DoRequest(ctx context.Context, Url string, Para Additional) (response *http.Response, err error) {

	//request, err := http.NewRequest(custom.Method, Url, nil)

	if Para.Mod == "default" {
		if !strings.HasPrefix(Para.Value, "/") {
			Para.Value = "/" + Para.Value
		}
		Url = Url + Para.Value
	}

	request, err := http.NewRequest("GET", Url, nil)

	if err != nil {
		return nil, err
	}

	request = request.WithContext(ctx)

	for _, header := range custom.Headers {
		request.Header.Set(header.Name, header.Value)
	}

	if Para.Mod == "host" {
		request.Header.Set("test", "test")
		request.Host = Para.Value
	}

	response, err = custom.CuClient.Do(request)

	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, fmt.Errorf("invalid certificate: %w", ue.Err)
			}
		}
		return nil, err
	}

	return response, nil

}
