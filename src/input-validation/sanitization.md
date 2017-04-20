Sanitization
============

Sanitization refers to the process of removing or replacing submitted data.
When dealing with data, after the proper validation checks have been made, an
additional step that is usually taken to strengthen data safety is sanitization.

The most common uses of sanitization are as follows:

## Convert single less-than characters `<` to entity

In the native package `html` there are two functions used for sanitization:
one for escaping HTML text and another for unescaping HTML.
The function `EscapeString()`, accepts a string and returns the same string
with the special characters escaped. i.e. `<` becomes `&lt;`.
Note that this function only escapes the following five characters: `<`, `>`,
`&`, `'` and `"`.
Conversely there is also the `UnescapeString()` function to convert from
entities to characters.

## Strip all tags

Although the `html/template` package has a `stripTags()` function, it's
unexported. Since no other native package has a function to strip all tags, the
alternatives are to use a third-party library, or to copy the whole function
along with it's private classes and functions.

Some of the third-party libraries availilable to achieve this are:

* https://github.com/kennygrant/sanitize
* https://github.com/maxwells/sanitize
* https://github.com/microcosm-cc/bluemonday

## Remove line breaks, tabs and extra white space

The `text/template` and the `html/template` include a way to remove
whitespaces from the template, by using a minus sign `-` inside the action's
delimiter.

Executing the template with source

```
{{- 23}} < {{45 -}}
```

will lead to the following output

```
23<45
```

**NOTE**: If the minus `-` sign is not placed immediately after the opening
action delimiter ``{{`` or before the closing action delimiter ``}}``, the
minus sign `-` will be applied to the value

Template source

```
{{ -3 }}
```

leads to

```
-3
```

## URL request path

In the `net/http` package there is an HTTP request multiplexer type called
`ServeMux`. It is used to match the incoming request to the registered
patterns, and calls the handler that most closely matches the
requested URL.
In addition to it's main purpose, it also takes care of sanitizing the URL
request path, redirecting any request containing `.` or `..` elements or
repeated slashes to an equivalent, cleaner URL.

A simple Mux example to illustrate:

```go
func main() {
  mux := http.NewServeMux()

  rh := http.RedirectHandler("http://yourDomain.org", 307)
  mux.Handle("/login", rh)

  log.Println("Listening...")
  http.ListenAndServe(":3000", mux)
}
```

Third-party packages:

* [Gorilla Toolkit - MUX][1]

[1]: http://www.gorillatoolkit.org/pkg/mux
