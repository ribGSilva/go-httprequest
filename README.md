# GO HttpRequest

[![Go Reference](https://pkg.go.dev/badge/github.com/ribGSilva/go-httprequest.svg)](https://pkg.go.dev/github.com/ribGSilva/go-httprequest)

This lib is an easy and re-utilizable way to make http requests

## Install

To install just run:

```ssh
    go get github.com/ribGSilva/go-httprequest
```

## Request

To make requests, follow the example:

```go
func main() {
    builder := rq.NewBuilder("my.host.com",
        rq.Method(http.MethodPost), // by default is GET
        rq.Path("/path/:id"),
        rq.Param("id", id),
        rq.Query("myQuery", "someValue"),
        rq.Header("Authorization", "myauth"),
        rq.JSON(body),
    )
	
	resp := rq.Exec[myType](builder)
	
	fmt.Printf("%+v", resp)
}
```

Developer:
Gabriel Ribeiro Silva
https://www.linkedin.com/in/gabriel-ribeiro-silva/