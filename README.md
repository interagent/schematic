# Schematic

Generate Go client code for HTTP APIs described by [JSON Schemas](http://json-schema.org/).

## Installation

Download and install:

```console
$ go get github.com/interagent/schematic
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
h := heroku.NewService(nil)
addons, err := h.AddonList("schematic", nil)
if err != nil {
  ...
}
for _, addon := range addons {
  fmt.Println(addon.Name)
}
```

A Service takes an optional ``http.Client`` as argument.
As an example, if your service use
[OAuth2](http://code.google.com/p/goauth2/) for authentication:

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

## Development

Schematic bundles templated Go code into a Go source file via the
[templates package](https://github.com/cyberdelia/templates). To rebuild
the Go source file after changing .tmpl files:

```console
$ templates -source schema/templates/ > schema/templates/templates.go
```
