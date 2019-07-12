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
* _Checks forpath alteration characters_ - `../` or `\\..`
* _Checks for Extended UTF-8_ - check for alternative representations of
  special characters

**Note**: Ensure that the HTTP request and response headers only contain
ASCII characters.

Third-party packages exist that handle security in Go:

* [Gorilla][6] - One of the most used packages for web
  application security.
  It has support for `websockets`, `cookie sessions`, `RPC`, among
  others.
* [Form][7] - Decodes `url.Values` into Go value(s) and Encodes Go value(s)
  into `url.Values`.
  Dual `Array` and Full `map` support.
* [Validator][8] - Go `Struct` and `Field` validation, including `Cross Field`,
  `Cross Struct`, `Map` as well as `Slice` and `Array` diving.

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

Anytime data is passed from a trusted source to a less-trusted source, integrity
checks should be made. This guarantees that the data has not been tampered with
and we are receiving the intended data. Other data source checks include:

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
