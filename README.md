
# routing

> **DEPRECATED:** Use https://github.com/altipla-consulting/libs instead.

[![GoDoc](https://godoc.org/github.com/altipla-consulting/routing?status.svg)](https://godoc.org/github.com/altipla-consulting/routing)
[![Build Status](https://travis-ci.org/altipla-consulting/routing.svg?branch=master)](https://travis-ci.org/altipla-consulting/routing)

Routing requests to handlers.


### Install

```shell
go get github.com/altipla-consulting/routing
```

This library depends on the following ones:
- [github.com/altipla-consulting/sentry](github.com/altipla-consulting/sentry)
- [github.com/julienschmidt/httprouter](github.com/julienschmidt/httprouter)
- [github.com/sirupsen/logrus](github.com/sirupsen/logrus)


### Usage

```go
package main

import (
  "fmt"
  "net/http"

  "github.com/altipla-consulting/routing"
  "github.com/altipla-consulting/langs"
  "github.com/julienschmidt/httprouter"
)

func RobotsHandler(w http.ResponseWriter, r *http.Request) error {
  fmt.Fprintln(w, "ok")
  return nil
}

func main {
  s := routing.NewServer(r)
  s.Get(langs.ES, "/robots.txt", RobotsHandler)
}
```


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using `gofmt`.


### Running tests

Run the tests

```shell
make test
```


### License

[MIT License](LICENSE)
