// Package httprequest package brings facilities to build http.Request.
// it brings re-utilizable codes with options with the most usual necessities
package httprequest

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Client for the request
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// Builder carries all the data necessary to execute a http request
type Builder struct {
	// Context for the request
	Context context.Context
	// Client to execute the request
	// Can use http.DefaultClient
	Client Client
	// method is the http GET, POST...
	Method string
	// Host is the host of the Builder
	// Example:
	// 		http://my.host.com
	Host string
	// Path is the path for the Builder
	// Example:
	//		/my/path
	//		/:myParam
	Path string
	// Params has the params to bind in the path
	Params map[string]string
	// Headers has the headers of the Builder
	Headers http.Header
	// Queries has the queries of the Builder
	Queries url.Values
	// Encoder has the encoder for the Body
	Encoder EncoderFunc
	// Body has the body for the Builder
	Body any
	// Decoder has the decoder for the response
	Decoder DecoderFunc
}

//EncoderFunc encodes the Body
type EncoderFunc func(any) ([]byte, error)

//DecoderFunc decodes the http request
type DecoderFunc func([]byte, any) error

// Option add optional values to the Builder
type Option func(*Builder)

// NewBuilder a new Builder
// Example:
//		func reqBuilder(ctx context.Context, id string, body any) {
//			builder := NewBuilder("http://my.host.com",
//				Method(MethodPatch), // by default is GET
//				Path("/path/:id"),
//				Param("id", id),
//				Query("myQuery", "someValue"),
//				Header("Authorization", "myauth"),
//				Body(body),
//			)
//		}
func NewBuilder(host string, options ...Option) *Builder {
	r := Builder{
		Context: context.Background(),
		Client:  http.DefaultClient,
		Method:  http.MethodGet,
		Host:    host,
		Params:  make(map[string]string),
		Headers: make(http.Header),
		Queries: make(url.Values),
		Encoder: json.Marshal,
		Decoder: json.Unmarshal,
	}
	for _, o := range options {
		o(&r)
	}

	return &r
}

func (b *Builder) Build() (*http.Request, error) {
	p := b.Path
	for k, v := range b.Params {
		p = strings.ReplaceAll(p, ":"+k, v)
	}

	base := fmt.Sprintf("%s%s", b.Host, p)

	var body io.Reader
	if b.Body != nil {
		b, err := b.Encoder(b.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(b.Context, b.Method, base, body)
	if err != nil {
		return nil, err
	}

	req.Header = b.Headers
	req.URL.RawQuery = b.Queries.Encode()

	return req, nil
}

// Response holds the data of the http response
type Response[T any] struct {
	Status           int
	Body             T
	Err              error
	OriginalResponse *http.Response
}

// Do performs a request and retrieves the response for the request
func Do[T any](b Builder) Response[T] {
	request, err := b.Build()
	if err != nil {
		return Response[T]{
			Err: err,
		}
	}

	response, err := b.Client.Do(request)
	if err != nil {
		return Response[T]{
			Err: err,
		}
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return Response[T]{
			Status:           response.StatusCode,
			OriginalResponse: response,
		}
	}

	body, err := ParseResponse[T](response, b.Decoder)

	return Response[T]{
		Status:           response.StatusCode,
		Body:             body,
		Err:              err,
		OriginalResponse: response,
	}
}

// ParseResponse parses the response into a struct
func ParseResponse[T any](r *http.Response, f DecoderFunc) (T, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	var body T
	if len(buf) > 0 {
		err := f(buf, &body)
		return body, err
	}
	return body, nil
}

// Ctx specify the context for the Builder
func Ctx(context context.Context) Option {
	return func(r *Builder) {
		r.Context = context
	}
}

// Cli specify the client to execute the request
func Cli(cli Client) Option {
	return func(r *Builder) {
		r.Client = cli
	}
}

// Method specify the http method for the Builder
func Method(method string) Option {
	return func(r *Builder) {
		r.Method = method
	}
}

// Path sets the path
// To set path params, use :{value}
// Example:
// 			...
// 			Path("/:userId/address/:addId")
//			Param("userId", "123")
//			Param("addId", "2")
// 			...
func Path(path string) Option {
	return func(r *Builder) {
		r.Path = path
	}
}

// Param adds a param bind
func Param(key string, value interface{}) Option {
	return func(r *Builder) {
		r.Params[key] = fmt.Sprint(value)
	}
}

// Params sets the params
func Params(params map[string]interface{}) Option {
	return func(r *Builder) {
		for k, v := range params {
			r.Params[k] = fmt.Sprint(v)
		}
	}
}

// Header adds to the header a value
// The header name will always be first letter Upper
// Example:
// 			...
// 			WithHeader("authoRIZATION", "someHASH")
// 			WithHeader("content-tyPE", "someContent")
// 			...
//     this will end up as a header:
//			Authorization: someHASH
//			Content-Type:  someContent
func Header(key string, value interface{}) Option {
	return func(r *Builder) {
		r.Headers.Add(key, fmt.Sprint(value))
	}
}

// Headers sets the headers
func Headers(headers http.Header) Option {
	return func(r *Builder) {
		r.Headers = headers
	}
}

// Query adds query param to the Builder
func Query(key string, value interface{}) Option {
	return func(r *Builder) {
		r.Queries.Add(key, fmt.Sprint(value))
	}
}

// Queries sets the query params
func Queries(queries url.Values) Option {
	return func(r *Builder) {
		r.Queries = queries
	}
}

// Encoder sets the encoder
func Encoder(f EncoderFunc) Option {
	return func(r *Builder) {
		r.Encoder = f
	}
}

// Decoder sets the decoder
func Decoder(f EncoderFunc) Option {
	return func(r *Builder) {
		r.Encoder = f
	}
}

// Body sets the body
func Body(body any) Option {
	return func(r *Builder) {
		r.Body = body
	}
}

// String sets the body as a string
func String(body string) Option {
	return func(r *Builder) {
		r.Body = bytes.NewBufferString(body)
		r.Encoder = func(any) ([]byte, error) {
			return []byte(body), nil
		}
	}
}

// JSON sets the body as a json
// This method already sets the Content-Type header as application/json
func JSON(body interface{}) Option {
	return func(r *Builder) {
		r.Body = body
		r.Encoder = json.Marshal
		r.Headers.Add("Content-Type", "application/json")
	}
}

// XML sets the body as a xml
// This method already sets the Content-Type header as application/xml
func XML(body interface{}) Option {
	return func(r *Builder) {
		r.Body = body
		r.Encoder = xml.Marshal
		r.Headers.Add("Content-Type", "application/xml")
	}
}
