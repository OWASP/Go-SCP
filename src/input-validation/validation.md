Validation
==========

In validation checks, the user input is checked against a set of conditions in
order to guarantee that the user is indeed entering the expected data.

**IMPORTANT:** If the validation fails, the input must be rejected.

This is important not only from a security standpoint but from the perspective
of data consistency and integrity, since data is usually used across a variety
of systems and applications.

This article lists the security risks developers should be aware of when
developing web applications in Go.

## User Interactivity

Any part of an application that allows user input is a potential security risk.
Problems can occur not only from threat actors that seek a way to compromise the
application, but also from erroneous input caused by human error (statistically,
the majority of the invalid data situations are usually caused by human error).
In Go there are several ways to protect against such issues.

Go has native libraries which include methods to help ensure such errors are
not made. When dealing with strings we can use packages like the following
examples:

* `strconv` package handles string conversion to other datatypes.
    * [`Atoi`](https://golang.org/pkg/strconv/#Atoi)
    * [`ParseBool`](https://golang.org/pkg/strconv/#ParseBool)
    * [`ParseFloat`](https://golang.org/pkg/strconv/#ParseFloat)
    * [`ParseInt`](https://golang.org/pkg/strconv/#ParseInt)
* `strings` package contains all functions that handle strings and its
  properties.
    * [`Trim`](https://golang.org/pkg/strings/#Trim)
    * [`ToLower`](https://golang.org/pkg/strings/#ToLower)
    * [`ToTitle`](https://golang.org/pkg/strings/#ToTitle)
* [`regexp`][4] package support for regular expressions to accommodate custom
   formats[^1].
* [`utf8`][9] package implements functions and constants to support text
  encoded in UTF-8. It includes functions to translate between runes and UTF-8
  byte sequences.

  Validating UTF-8 encoded runes:
    * [`Valid`](https://golang.org/pkg/unicode/utf8/#Valid)
    * [`ValidRune`](https://golang.org/pkg/unicode/utf8/#ValidRune)
    * [`ValidString`](https://golang.org/pkg/unicode/utf8/#ValidString)

  Encoding UTF-8 runes:
    * [`EncodeRune`](https://golang.org/pkg/unicode/utf8/#EncodeRune)

  Decoding UTF-8:
    * [`DecodeLastRune`](https://golang.org/pkg/unicode/utf8/#DecodeLastRune)
    * [`DecodeLastRuneInString`](https://golang.org/pkg/unicode/utf8/#DecodeLastRuneInString)
    * [`DecodeRune`](https://golang.org/pkg/unicode/utf8/#DecodeLastRune)
    * [`DecodeRuneInString`](https://golang.org/pkg/unicode/utf8/#DecodeRuneInString)


**Note**: `Forms` are treated by Go as `Maps` of `String` values.

Other techniques to ensure the validity of the data include:

* _Whitelisting_ - whenever possible validate the input against a whitelist
  of allowed characters. See [Validation - Strip tags][1].
* _Boundary checking_ - both data and numbers length should be verified.
* _Character escaping_ - for special characters such as standalone quotation
  marks.
* _Numeric validation_ - if input is numeric.
* _Check for Null Bytes_ - `(%00)`
* _Checks for new line characters_ - `%0d`, `%0a`, `\r`, `\n`
* _Checks for path alteration characters_ - `../` or `\\..`
* _Checks for Extended UTF-8_ - check for alternative representations of
  special characters

**Note**: Ensure that the HTTP request and response headers only contain
ASCII characters.

Third-party packages exist that handle security in Go:

* [Gorilla][6] - One of the most used packages for web
  application security. (Note: was archived Dec 2022, revived Jul 2023)
  It has support for `websockets`, `cookie sessions`, `RPC`, among
  others.
* [Form][7] - Decodes `url.Values` into Go value(s) and Encodes Go value(s)
  into `url.Values`.
  Dual `Array` and Full `map` support.
* [Validator][8] - Go `Struct` and `Field` validation, including `Cross Field`,
  `Cross Struct`, `Map` as well as `Slice` and `Array` diving.

## Using go-playground/validator (Recommended)

The `go-playground/validator` package is the de-facto standard for struct
validation in Go. It provides declarative validation via struct tags.

### Installation

```bash
go get github.com/go-playground/validator/v10
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/go-playground/validator/v10"
)

type UserRegistration struct {
    Username string `validate:"required,alphanum,min=3,max=32"`
    Email    string `validate:"required,email,max=255"`
    Password string `validate:"required,min=8,max=72"` // 72 = bcrypt limit
    Age      int    `validate:"gte=13,lte=120"`
    Website  string `validate:"omitempty,url"`
    Role     string `validate:"required,oneof=user admin moderator"`
}

var validate = validator.New()

func validateUser(user *UserRegistration) error {
    err := validate.Struct(user)
    if err != nil {
        // Don't expose detailed validation errors to users
        // Log them server-side for debugging
        for _, e := range err.(validator.ValidationErrors) {
            fmt.Printf("Field %s failed %s validation\n", e.Field(), e.Tag())
        }
        return fmt.Errorf("invalid input")
    }
    return nil
}
```

### Common Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must not be zero value | `required` |
| `email` | Valid email format | `email` |
| `url` | Valid URL | `url` |
| `min` | Minimum length/value | `min=8` |
| `max` | Maximum length/value | `max=255` |
| `len` | Exact length | `len=10` |
| `gte` | Greater than or equal | `gte=0` |
| `lte` | Less than or equal | `lte=100` |
| `oneof` | Must be one of values | `oneof=red green blue` |
| `alphanum` | Alphanumeric only | `alphanum` |
| `uuid` | Valid UUID | `uuid` |
| `ip` | Valid IP address | `ip` |
| `cidr` | Valid CIDR notation | `cidr` |

### Security-Focused Custom Validators

```go
func init() {
    // Safe filename validator (no path traversal)
    validate.RegisterValidation("safe_filename", func(fl validator.FieldLevel) bool {
        filename := fl.Field().String()
        // Reject path separators and traversal
        return !strings.ContainsAny(filename, "/\\") &&
               filename != "." && filename != ".." &&
               filepath.IsLocal(filename)
    })

    // No HTML tags validator (XSS prevention)
    validate.RegisterValidation("no_html", func(fl validator.FieldLevel) bool {
        return !strings.ContainsAny(fl.Field().String(), "<>")
    })

    // Alphanumeric with limited special chars
    validate.RegisterValidation("safe_string", func(fl validator.FieldLevel) bool {
        matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_\.]+$`, fl.Field().String())
        return matched
    })
}

type FileUpload struct {
    Filename string `validate:"required,safe_filename,max=255"`
    Comment  string `validate:"omitempty,no_html,max=1000"`
}
```

### Cross-Field Validation

```go
type PasswordChange struct {
    CurrentPassword string `validate:"required,min=8"`
    NewPassword     string `validate:"required,min=8,nefield=CurrentPassword"`
    ConfirmPassword string `validate:"required,eqfield=NewPassword"`
}
```

### Slice and Map Validation

```go
type BulkOperation struct {
    // Validate each item in the slice
    IDs []int64 `validate:"required,min=1,max=100,dive,gt=0"`

    // Validate map keys and values
    Tags map[string]string `validate:"max=10,dive,keys,alphanum,max=32,endkeys,max=100"`
}
```

### HTTP Handler Integration

```go
func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var input UserRegistration

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if err := validate.Struct(&input); err != nil {
        // Log detailed error server-side
        log.Printf("Validation failed: %v", err)

        // Return generic error to client (don't leak field names)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Input is valid, proceed with business logic
    // ...
}
```

### Validation Error Messages (Secure Pattern)

```go
func formatValidationErrors(err error) map[string]string {
    errors := make(map[string]string)

    for _, e := range err.(validator.ValidationErrors) {
        // Map technical tags to user-friendly messages
        switch e.Tag() {
        case "required":
            errors[e.Field()] = "This field is required"
        case "email":
            errors[e.Field()] = "Invalid email format"
        case "min":
            errors[e.Field()] = fmt.Sprintf("Must be at least %s characters", e.Param())
        case "max":
            errors[e.Field()] = fmt.Sprintf("Must be at most %s characters", e.Param())
        default:
            errors[e.Field()] = "Invalid value"
        }
    }

    return errors
}
```

**Important Security Notes:**

1. **Validation is not sanitization** - Always sanitize output even if input
   is validated
2. **Don't trust validation alone** - Use parameterized queries, output
   encoding, etc.
3. **Validate at boundaries** - Validate data entering your system, not just
   at the edge
4. **Log validation failures** - They may indicate attack attempts

## File Manipulation

Any time file usage is required ( `read` or `write` a file ), validation checks
should also be performed, since most of the file manipulation operations deal
with user data.

Other file check procedures include "File existence check", to verify that a
filename exists.

Addition file information is in the [File Management][2] section and information
regarding `Error Handling` can be found in the [Error Handling][3] section of
the document.

## Data sources

Anytime data is passed from a less-trusted source to a more-trusted source (for
example, user input, external APIs, or client requests to your application),
integrity checks and validation should be performed. This ensures that the data
has not been tampered with and meets the expected format and constraints. Other
data source checks include:

* _Cross-system consistency checks_
* _Hash totals_
* _Referential integrity_

**Note:** In modern relational databases, if values in the primary key field
are not constrained by the database's internal mechanisms then they should be
validated.

* _Uniqueness check_
* _Table look up check_

## Post-validation Actions

According to Data Validation's best practices, the input validation is only
the first part of the data validation guidelines. Therefore,
_Post-validation Actions_ should also be performed.
The _Post-validation Actions_ used vary with the context and are divided in
three separate categories:

* **Enforcement Actions**
  Several types of _Enforcement Actions_ exist in order to better secure our
  application and data.

  * inform the user that submitted data has failed to comply with the
    requirements and therefore the data should be modified in order to comply
    with the required conditions.
  * modify user submitted data on the server side without notifying the user of
    the changes made. This is most suitable in systems with interactive usage.

  **Note:** The latter is used mostly in cosmetic changes (modifying sensitive
  user data can lead to problems like truncating, which result in data loss).
* **Advisory Action**
  Advisory Actions usually allow for unchanged data
  to be entered, but the source actor is informed that there were issues with
  said data. This is most suitable for non-interactive systems.
* **Verification Action**
  Verification Action refer to special cases in Advisory Actions. In these
  cases, the user submits the data and the source actor asks the user to verify
  the data and suggests changes. The user then accepts these changes or keeps
  his original input.

  A simple way to illustrate this is a billing address form, where the user
  enters his address and the system suggests addresses associated with the
  account. The user then accepts one of these suggestions or ships to the
  address that was initially entered.

---

[^1]: Before writing your own regular expression have a look at [OWASP Validation Regex Repository][5]

[1]: sanitization.md
[2]: ../file-management/README.md
[3]: ../error-handling-logging/README.md
[4]: https://golang.org/pkg/regexp/
[5]: https://www.owasp.org/index.php/OWASP_Validation_Regex_Repository
[6]: https://github.com/gorilla/
[7]: https://github.com/go-playground/form
[8]: https://github.com/go-playground/validator
[9]: https://golang.org/pkg/unicode/utf8/
