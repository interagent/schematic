# Schematic

Generate Go client code for HTTP APIs described by [JSON Hyper-Schemas](http://json-schema.org/latest/json-schema-hypermedia.html).

## Installation

Download and install:

```console
$ go get -u github.com/interagent/schematic
```

**Warning**: schematic requires Go >= 1.2.

## Usage

Run it against your schema:

```console
$ schematic platform-api.json > heroku/heroku.go
```

This will generate a Go package named after your schema:

```go
package heroku
...
```

You then would be able to use the package as follow:

```go
package main

import "./heroku"

func main() {
  h := heroku.NewService(nil)
  addons, err := h.AddonServiceList(nil)
  if err != nil {
    panic(err)
  }
  for _, addon := range addons {
    fmt.Println(addon.Name)
  }
}
```

A Service takes an optional [`http.Client`](http://golang.org/pkg/net/http/#Client)
as argument. This Client will need to handle any HTTP-related details
unique to the target API service such as request authorization, request
header setup, error handling, response debugging, etc.

As an example, if your service use [OAuth2](http://code.google.com/p/goauth2/)
for authentication:

```go
var config = &oauth.Config{
  ...
}
transport := &oauth.Transport{
  Token:     token,
  Config:    config,
  Transport: http.DefaultTransport,
}
httpClient := transport.Client()
s := api.NewService(httpClient)
```

For an example of a service using HTTP Basic Auth, see the generated
[heroku-go client](https://github.com/cyberdelia/heroku-go/blob/master/v3/transport.go).

In general, creating a custom [`http.Transport`](http://golang.org/pkg/net/http/#Transport)
and creating a client from that is a common way to get an appropriate
`http.Client` for your service.

## Development

Schematic bundles templated Go code into a Go source file via the
[templates package](https://github.com/cyberdelia/templates). To rebuild
the Go source file after changing .tmpl files:

```console
$ templates -source schema/templates/ > schema/templates/templates.go
```
