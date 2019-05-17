[![This project is considered experimental](https://img.shields.io/badge/Status-experimental-red.svg)](https://arp242.net/status/experimental)
[![Build Status](https://travis-ci.org/Carpetsmoker/mhttp.svg?branch=master)](https://travis-ci.org/Carpetsmoker/mhttp)
[![GoDoc](https://godoc.org/github.com/Carpetsmoker/mhttp?status.svg)](https://godoc.org/github.com/Carpetsmoker/mhttp)

Martin's HTTP package. Contains various things I find useful and copy/pasted
once too many times.

<!--
- `mhttp.Wrap()` – allow returning errors from HTTP endpoints:

      http.HandleFunc("/bar", mhttp.Wrap(func(w http.ResponseWriter, r *http.Request) error {
          d, err := getData()
          if err != nil {
              return err
          }

          return mhttp.String("Hello, %s", d)
      }))

  It's just more convenient than `http.Error(...)` followed by a `return`.

  Return helpers:

  - `mhttp.String()`
  - `mhttp.Template()`
  - `mhttp.SeeOther()`

- Middlewares:

  - `mtthp.Headers()` – Set HTTP headers on requests.
  - `mhttp.Log()`     – Log requests, mostly intended for dev.
  - `mhttp.Unpanic()` – Handle panics.
  - `mhttp.Auth()`    – Authentication.

- Template loading/reloading:

- Flash messages:

-->
