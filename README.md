Martin's HTTP package: It's not a "framework", but just a collection of
functions for building HTTP services.

Honestly, I'm not sure how useful this will be for other people at this stage,
but it's pretty useful for me.

Much of this was extracted from
[GoatCounter](https://github.com/zgoat/goatcounter).

---

`zhttp.Wrap()` allows returning errors from HTTP endpoints:

```go
http.HandleFunc("/bar", zhttp.Wrap(func(w http.ResponseWriter, r *http.Request) error {
    d, err := getData()
    if err != nil {
        return err
    }

    return zhttp.Text(w, "Hello, %s", d)
}))
```

It's just more convenient than `http.Error(...)` followed by a `return`. The
`ErrFunc()` will be used to report returned errors (you can override it if you
need to).

Return helpers (with the `http.ResponseWriter` elided in the function
signature):

    Stream(fp io.Reader)             Stream any data.
    Bytes(b []byte)                  Send []byte
    String(s string)                 Send string
    Text(s string)                   Send string with Content-Type text/plain
    JSON(i any)                      Send JSON
    Template(name string, data any)  Render a template (see below)
    MovedPermanently(url string)     301 Moved Permanently
    SeeOther(url string)             303 See Other

---

Templates can be rendered with the `ztpl` package; you have to call
`ztpl.Init()` first with the path to load templates from, and an optional map
for templates compiled in the binary.

You can use `ztpl.Reload()` to reload the templates from disk on changes, which
is useful for development. e.g. with github.com/teamwork/reload:


```go
ztpl.Init("tpl", pack.Templates)

go func() {
    err := reload.Do(zlog.Module("main").Debugf, reload.Dir("./tpl", ztpl.Reload))
    if err != nil {
        panic(errors.Errorf("reload.Do: %v", err))
    }
}()
```


---

Few other tidbits:

- User auth can be added with `zhttpAuth()`, `zhttp.SetAuthCookie()`, and
  `zhttp.Filter()`.

- `zhttp.Decode()`scans forms, JSON body, or URL query parameters in to a
  struct. It's just a convencience wrapper around formam.

- `zhttp.NewStatic()` will create a static file host.

- `zhttp.HostRoute()` routes request to chi routers based on the Host header.
