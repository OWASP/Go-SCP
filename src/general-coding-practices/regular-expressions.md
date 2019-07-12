Regular Expressions
===================

Regular Expressions are a powerful tool that's widely used to perform searches
and validations. In the context of a web applications they are commonly used to
perform input validation (e.g. Email address).

> Regular expressions are a notation for describing sets of character strings.
> When a particular string is in the set described by a regular expression, we
> often say that the regular expression matches the string. ([source][1])

It is well-known that Regular Expressions are hard to master. Sometimes, what
seems to be a simple validation, may lead to a [Denial-of-Service][2].

Go authors took it seriously, and unlike other programming languages, the
decided to implement [RE2][3] for the [regex standard package][4].

## Why RE2

> RE2 was designed and implemented with an explicit goal of being able to handle
> regular expressions from untrusted users without risk. ([source][10])

With security in mind, RE2 also guarantees a linear-time performance and
graceful failing: the memory available to the parser, the compiler, and the
execution engines is limited.

## Regular Expression Denial of Service (ReDoS)

> Regular Expression Denial of Service (ReDoS) is an algorithmic complexity
> attack that provokes a Denial of Service (DoS). ReDos attacks are caused by a
> regular expression that takes a very long time to be evaluated, exponentially
> related with the input size. This exceptionally long time in the evaluation
> process is due to the implementation of the regular expression in use, for
> example, recursive backtracking ones. ([source][8])

You're better off reading the full article "[Diving Deep into Regular Expression
Denial of Service (ReDoS) in Go][8]" as it goes deep into the problem, and also
includes comparisons between the most popular programming languages. In this
section we will focus on a real-world use case.

Say for some reason you're looking for a Regular Expression to validate Email
addresses provided on your signup form. After a quick search, you found this
[RegEx for email validation at RegExLib.com][9]:

```
^([a-zA-Z0-9])(([\-.]|[_]+)?([a-zA-Z0-9]+))*(@){1}[a-z0-9]+[.]{1}(([a-z]{2,3})|([a-z]{2,3}[.]{1}[a-z]{2,3}))$
```

If you try to match `john.doe@somehost.com` against this regular expression you
may feel confident that it does what you're looking for. If you're developing
using Go, you'll come up with something like this:

```go
package main

import (
    "fmt"
    "regexp"
)

func main() {
    testString1 := "john.doe@somehost.com"
    testString2 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!"
    regex := regexp.MustCompile("^([a-zA-Z0-9])(([\\-.]|[_]+)?([a-zA-Z0-9]+))*(@){1}[a-z0-9]+[.]{1}(([a-z]{2,3})|([a-z]{2,3}[.]{1}[a-z]{2,3}))$")

    fmt.Println(regex.MatchString(testString1))
    // expected output: true
    fmt.Println(regex.MatchString(testString2))
    // expected output: false
}
```

Which is not a problem:

```
$ go run src/redos.go
true
false
```

However, what if you're developing with, for example, JavaScript?

```JavaScript
const testString1 = 'john.doe@somehost.com';
const testString2 = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!';
const regex = /^([a-zA-Z0-9])(([\-.]|[_]+)?([a-zA-Z0-9]+))*(@){1}[a-z0-9]+[.]{1}(([a-z]{2,3})|([a-z]{2,3}[.]{1}[a-z]{2,3}))$/;

console.log(regex.test(testString1));
// expected output: true
console.log(regex.test(testString2));
// expected output: hang/FATAL EXCEPTION

```

In this case, **execution will hang forever** and your application will service
no further requests (at least this process). This means **no further signups
will work until the application gets restarted**, resulting in **business
losses**.

## What's missing?

If you have a background with other programming languages such as Perl, Python,
PHP, or JavaScript, you should be aware of the differences regarding Regular
Expression supported features.

RE2 does not support constructs where only backtracking solutions are known to
exist, such as [Backreferences][5] and [Lookaround][6].

Consider the following problem: validating whether an arbitrary string is a
well-formed HTML tag: a) opening and closing tag names match, and b) optionally
there's some text in between.

Fulfilling requirement b) is straightforward `.*?`. But fulling requirement a)
is challenging because closing a tag match depends on what was matched as the
opening tag. This is exactly what Backreferences allows us to do. See the
JavaScript implementation below:

```JavaScript
const testString1 = '<h1>Go Secure Coding Practices Guide</h1>';
const testString2 = '<p>Go Secure Coding Practices Guide</p>';
const testString3 = '<h1>Go Secure Coding Practices Guid</p>';
const regex = /<([a-z][a-z0-9]*)\b[^>]*>.*?<\/\1>/;

console.log(regex.test(testString1));
// expected output: true
console.log(regex.test(testString2));
// expected output: true
console.log(regex.test(testString3));
// expected output: false

```

`\1` will hold the value previously captured by `([A-Z][A-Z0-9]*)`.

This is something you should not expect to do in Go.

```go
package main

import (
    "fmt"
    "regexp"
)

func main() {
    testString1 := "<h1>Go Secure Coding Practices Guide</h1>"
    testString2 := "<p>Go Secure Coding Practices Guide</p>"
    testString3 := "<h1>Go Secure Coding Practices Guid</p>"
    regex := regexp.MustCompile("<([a-z][a-z0-9]*)\b[^>]*>.*?<\/\1>")

    fmt.Println(regex.MatchString(testString1))
    fmt.Println(regex.MatchString(testString2))
    fmt.Println(regex.MatchString(testString3))
}

```

Running the Go source code sample above should result in the following errors:

```
$ go run src/backreference.go
# command-line-arguments
src/backreference.go:12:64: unknown escape sequence
src/backreference.go:12:67: non-octal character in escape sequence: >
```

You may feel tempted to fix these errors, coming up with the following regular
expression:

```
<([a-z][a-z0-9]*)\b[^>]*>.*?<\\/\\1>
```

Then, this is what you'll get:

```
go run src/backreference.go
panic: regexp: Compile("<([a-z][a-z0-9]*)\b[^>]*>.*?<\\/\\1>"): error parsing regexp: invalid escape sequence: `\1`

goroutine 1 [running]:
regexp.MustCompile(0x4de780, 0x21, 0xc00000e1f0)
        /usr/local/go/src/regexp/regexp.go:245 +0x171
main.main()
        /go/src/backreference.go:12 +0x3a
exit status 2
```

While developing something from scratch, you'll probably find a nice workaround
to help with the lack of some features. On the other hand, porting existing
software could make you look for full featured alternative to the standard
Regular Expression package, and you'll likely find some (e.g.
[dlclark/regexp2][7]). Keeping that in mind, then you'll (probably) lose RE2's
"safety features" such as the linear-time performance.

[1]: https://swtch.com/~rsc/regexp/regexp1.html
[2]: #regular-expression-denial-of-service-redos
[3]: https://github.com/google/re2/wiki
[4]: https://golang.org/pkg/regexp/
[5]: https://www.regular-expressions.info/backref.html
[6]: https://www.regular-expressions.info/lookaround.html
[7]: https://github.com/dlclark
[8]: https://www.checkmarx.com/2018/05/07/redos-go/
[9]: http://regexlib.com/REDetails.aspx?regexp_id=1757
[10]: https://github.com/google/re2/wiki/WhyRE2
