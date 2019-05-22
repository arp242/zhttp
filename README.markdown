[![This project is considered experimental](https://img.shields.io/badge/Status-experimental-red.svg)](https://arp242.net/status/experimental)
[![Build Status](https://travis-ci.org/zgoat/zhttp.svg?branch=master)](https://travis-ci.org/zgoat/zhttp)
[![codecov](https://codecov.io/gh/zgoat/zhttp/branch/master/graph/badge.svg)](https://codecov.io/gh/zgoat/zhttp)
[![GoDoc](https://godoc.org/github.com/zgoat/zhttp?status.svg)](https://godoc.org/github.com/zgoat/zhttp)

Martin's HTTP package. Contains various things I find useful and copy/pasted
once too many times.

<!--
- `zhttp.Wrap()` – allow returning errors from HTTP endpoints:

      http.HandleFunc("/bar", zhttp.Wrap(func(w http.ResponseWriter, r *http.Request) error {
          d, err := getData()
          if err != nil {
              return err
          }

          return zhttp.String("Hello, %s", d)
      }))

  It's just more convenient than `http.Error(...)` followed by a `return`.

  Return helpers:

  - `zhttp.String()`
  - `zhttp.Template()`
  - `zhttp.SeeOther()`

- Middlewares:

  - `mtthp.Headers()` – Set HTTP headers on requests.
  - `zhttp.Log()`     – Log requests, mostly intended for dev.
  - `zhttp.Unpanic()` – Handle panics.
  - `zhttp.Auth()`    – Authentication.

- Template loading/reloading:

- Flash messages:

-->
