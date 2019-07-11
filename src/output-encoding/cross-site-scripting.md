XSS - Cross Site Scripting
==========================

Although most developers have heard about it, most have never tried to exploit
a Web Application using XSS.

Cross Site Scripting has been on [OWASP Top 10][0] security risks since 2003 and
it's still a common vulnerability. The [2013 version][1] is quite detailed about
XSS, for example: attack vectors, security weakness, technical impacts and
business impacts.

In short

> You are vulnerable if you do not ensure that all user supplied input is
> properly escaped, or you do not verify it to be safe via server-side input
> validation, before including that input in the output page. ([source][1])

Go, just like any other multi-purpose programming language, has everything
needed to mess with and make you vulnerable to XSS, despite the documentation
being clear about using the [html/template package][2]. Quite easily, you can
find "hello world" examples using [net/http][3] and [io][4] packages. And
without realizing it, you're vulnerable to XSS.

Imagine the following code:

```go
package main

import "net/http"
import "io"

func handler (w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, r.URL.Query().Get("param1"))
}

func main () {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

This snippet creates and starts an HTTP Server listening on port `8080`
(`main()`), handling requests on server's root (`/`).

The `handler()` function, which handles requests, expects a Query String
parameter `param1`, whose value is then written to the response stream (`w`).

As `Content-Type` HTTP response header is not explicitly defined, Go
`http.DetectContentType` default value will be used, which follows the
[WhatWG spec][5].

So, making `param1` equal to "test", will result in `Content-Type` HTTP
response header to be sent as `text/plain`.

![Content-Type: text/plain][content-type-text-plain]

But if `param1` first characters are "&lt;h1&gt;", `Content-Type` will be
`text/html`.

![Content-Type: text/html][content-type-text-html]

You may think that making `param1` equal to any HTML tag will lead to the same
behavior, but it won't. Making `param1` equal to "&lt;h2&gt;", "&lt;span&gt;"
or "&lt;form&gt;" will make `Content-Type` to be sent as `plain/text` instead
of expected `text/html`.

Now let's make `param1` equal to `<script>alert(1)</script>`.

As per the [WhatWG spec][5], `Content-Type` HTTP response header will be sent as
`text/html`, `param1` value will be rendered, and here it is, the XSS (Cross
Site Scripting).

![XSS - Cross-Site Scripting][cross-site-scripting]

After talking with Google regarding this situation, they informed us that:

> It's actually very convenient and intended to be able to print html and have
> the content-type set automatically. We expect that programmers will use
> html/template for proper escaping.

Google states that developers are responsible for sanitizing and protecting
their code. We totally agree BUT in a language where security is a priority,
allowing `Content-Type` to be set automatically besides having `text/plain` as
default, is not the best way to go.

Let's make it clear: `text/plain` and/or the [text/template package][6] won't
keep you away from XSS, since it does not sanitize user input.

```go
package main

import "net/http"
import "text/template"

func handler(w http.ResponseWriter, r *http.Request) {
        param1 := r.URL.Query().Get("param1")

        tmpl := template.New("hello")
        tmpl, _ = tmpl.Parse(`{{define "T"}}{{.}}{{end}}`)
        tmpl.ExecuteTemplate(w, "T", param1)
}

func main() {
        http.HandleFunc("/", handler)
        http.ListenAndServe(":8080", nil)
}
```

Making `param1` equal to "&lt;h1&gt;" will lead to `Content-Type` being sent as
`text/html`. This is what makes you vulnerable to XSS.

![XSS while using text/template package][text-template-xss]

By replacing the [text/template package][6] with the [html/template][2] one,
you'll be ready to proceed... safely.

```go
package main

import "net/http"
import "html/template"

func handler(w http.ResponseWriter, r *http.Request) {
        param1 := r.URL.Query().Get("param1")

        tmpl := template.New("hello")
        tmpl, _ = tmpl.Parse(`{{define "T"}}{{.}}{{end}}`)
        tmpl.ExecuteTemplate(w, "T", param1)
}

func main() {
        http.HandleFunc("/", handler)
        http.ListenAndServe(":8080", nil)
}
```

Not only `Content-Type` HTTP response header will be sent as `text/plain` when
`param1` is equal to "&lt;h1&gt;"

![Content-Type: text/plain while using html/template package][html-template-plain-text]

but also `param1` is properly encoded to the output media: the browser.

![No XSS while using html/template package][html-template-noxss]

[exploit-of-a-mom]: images/exploit-of-a-mom.png
[content-type-text-plain]: images/text-plain.png
[content-type-text-html]: images/text-html.png
[cross-site-scripting]: images/xss.png
[text-template-xss]: images/text-template-xss.png
[html-template-plain-text]: images/html-template-plain-text.png
[html-template-noxss]: images/html-template-text-plain-noxss.png

[0]: https://www.owasp.org/index.php/Category:OWASP_Top_Ten_Project
[1]: https://www.owasp.org/index.php/Top_10_2013-A3-Cross-Site_Scripting_(XSS)
[2]: https://golang.org/pkg/html/template/
[3]: https://golang.org/pkg/net/http/
[4]: https://golang.org/pkg/io/
[5]: https://mimesniff.spec.whatwg.org/#rules-for-identifying-an-unknown-mime-typ
[6]: https://golang.org/pkg/text/template/
