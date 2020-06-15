Access Control
==============

When dealing with access controls the first step to take is to use only trusted
system objects for access authorization decisions.
In the example provided in the [Session Management][3] section, we implemented
this using JWT: JSON Web Tokens to generate a session token on the server-side.

```go
// create a JWT and put in the clients cookie
func setToken(res http.ResponseWriter, req *http.Request) {
    //30m Expiration for non-sensitive applications - OWASP
    expireToken := time.Now().Add(time.Minute * 30).Unix()
    expireCookie := time.Now().Add(time.Minute * 30)

    //token Claims
    claims := Claims{
        {...}
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, _ := token.SignedString([]byte("secret"))
```

We can then store and use this token to validate the user and enforce our
`Access Control` model.

The component used for access authorization should be a single one, used
site-wide. This includes libraries that call external authorization services.

In case of a failure, access control should fail securely. In Go we can use
`defer` to achieve this.
There are more details in the [Error Logging][1] section of this document.

If the application cannot access its configuration information, all
access to the application should be denied.

Authorization controls should be enforced on every request, including
server-side scripts, as well as requests from client-side technologies like AJAX
or Flash.

It is also important to properly separate privileged logic from the rest of the
application code.

Other important operations where access controls must be enforced in order to
prevent an unauthorized user from accessing them, are as follows:

* File and other resources
* Protected URL's
* Protected functions
* Direct object references
* Services
* Application data
* User and data attributes and policy information.

In the provided example, a simple direct object reference is tested. This code
is built upon the [sample in the Session Management][2].

When implementing these access controls, it's important to verify that the
server-side implementation and the presentation layer representations of access
control rules are the same.

If _state data_ needs to be stored on the client-side, it's necessary to use
encryption and integrity checking in order to prevent tampering.

Application logic flow must comply with the business rules.

When dealing with transactions, the number of transactions a single user or
device can perform in a given period of time must be above the business
requirements but low enough to prevent a user from performing a
Denial-of-Service (DoS) attack.

It is important to note that using only the `referer` HTTP header is
insufficient to validate authorization, and should only be used as a
supplemental check.

Regarding long authenticated sessions, the application should periodically
re-evaluate the user's authorization to verify that the user's permissions
have not changed. If the permissions have changed, log the user "out" and force
them to re-authenticate.

User accounts should also have a way to audit them, in order to comply with
safety procedures. (e.g. Disabling a user's account 30 days after the
password's expiration date).

The application must also support the disabling of accounts and the termination
of sessions when a user's authorization is revoked. (e.g. role change,
employment status, etc.).

When supporting external service accounts and accounts that support connections
_from_ or _to_ external systems, these accounts must use the lowest level
privilege possible.

[1]: ../error-handling-logging/error-handling.md
[2]: URL.go
[3]: ../session-management/README.md
