# Schematic

## Installation

Download and install :

```
$ go get github.com/heroku/schematic
```

## Usage

Run it against your schema :

```
$ schematic api.json > heroku/heroku.go 
```

This will generate a Go package named after your schema:
```go
package heroku
...
```

You then would be able to use the package as follow :
```go
h := heroku.NewService(nil)
addons, err := h.AddonList("schematic", nil)
if err != nil {
  ...
}
for _, addon := range addons {
  fmt.Println(addons.Name)
}
```

You can also provide a custom client:

```go
h := heroku.NewService(&HerokuClient{})
```


