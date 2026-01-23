Passkeys and WebAuthn
=====================

Passkeys (based on the WebAuthn/FIDO2 standard) represent the future of
authentication. They replace passwords with cryptographic key pairs, providing
phishing-resistant authentication without the usability issues of passwords.

Why Passkeys?
-------------

| Feature | Passwords | Passkeys |
|---------|-----------|----------|
| Phishing resistance | No | Yes |
| Credential reuse | Common | Impossible |
| Server breaches | Expose hashes | No secrets stored |
| User friction | High | Low (biometric) |
| Brute force | Possible | Impossible |

How Passkeys Work
-----------------

1. **Registration**: User's device generates a public/private key pair. The
   public key is sent to the server; the private key never leaves the device.

2. **Authentication**: Server sends a challenge. The device signs it with the
   private key (after biometric/PIN verification). Server verifies the signature
   with the stored public key.

Implementation in Go
--------------------

The [go-webauthn/webauthn][1] library provides WebAuthn support for Go
applications.

### Installation

```bash
go get github.com/go-webauthn/webauthn/webauthn
```

### Basic Setup

```go
package main

import (
    "github.com/go-webauthn/webauthn/webauthn"
)

var webAuthn *webauthn.WebAuthn

func init() {
    var err error
    webAuthn, err = webauthn.New(&webauthn.Config{
        RPDisplayName: "Your Application",    // Display name
        RPID:          "example.com",         // Your domain
        RPOrigins:     []string{"https://example.com"}, // Allowed origins
    })
    if err != nil {
        panic(err)
    }
}

// User must implement webauthn.User interface
type User struct {
    ID          []byte
    Name        string
    DisplayName string
    Credentials []webauthn.Credential
}

func (u *User) WebAuthnID() []byte                         { return u.ID }
func (u *User) WebAuthnName() string                       { return u.Name }
func (u *User) WebAuthnDisplayName() string                { return u.DisplayName }
func (u *User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
```

### Registration Flow

```go
// Step 1: Begin registration (send to client)
func beginRegistration(w http.ResponseWriter, r *http.Request) {
    user := getCurrentUser(r) // Get user from session

    options, session, err := webAuthn.BeginRegistration(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Store session data (use secure session storage)
    storeWebAuthnSession(r, session)

    // Send options to client
    json.NewEncoder(w).Encode(options)
}

// Step 2: Complete registration (receive from client)
func finishRegistration(w http.ResponseWriter, r *http.Request) {
    user := getCurrentUser(r)
    session := getWebAuthnSession(r)

    credential, err := webAuthn.FinishRegistration(user, *session, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Store credential for user in database
    user.Credentials = append(user.Credentials, *credential)
    saveUser(user)

    w.WriteHeader(http.StatusOK)
}
```

### Authentication Flow

```go
// Step 1: Begin authentication (send to client)
func beginLogin(w http.ResponseWriter, r *http.Request) {
    user := getUserByUsername(r.URL.Query().Get("username"))

    options, session, err := webAuthn.BeginLogin(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    storeWebAuthnSession(r, session)
    json.NewEncoder(w).Encode(options)
}

// Step 2: Complete authentication (receive from client)
func finishLogin(w http.ResponseWriter, r *http.Request) {
    user := getUserByUsername(r.URL.Query().Get("username"))
    session := getWebAuthnSession(r)

    credential, err := webAuthn.FinishLogin(user, *session, r)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }

    // Update credential sign count (prevents cloning attacks)
    updateCredentialSignCount(user, credential)

    // Create authenticated session
    createUserSession(w, r, user)

    w.WriteHeader(http.StatusOK)
}
```

### Client-Side JavaScript

```javascript
// Registration
async function registerPasskey() {
    // Get options from server
    const optionsResponse = await fetch('/webauthn/register/begin');
    const options = await optionsResponse.json();

    // Create credential (triggers biometric prompt)
    const credential = await navigator.credentials.create({
        publicKey: options.publicKey
    });

    // Send credential to server
    await fetch('/webauthn/register/finish', {
        method: 'POST',
        body: JSON.stringify(credential),
        headers: { 'Content-Type': 'application/json' }
    });
}

// Authentication
async function loginWithPasskey(username) {
    const optionsResponse = await fetch(`/webauthn/login/begin?username=${username}`);
    const options = await optionsResponse.json();

    const assertion = await navigator.credentials.get({
        publicKey: options.publicKey
    });

    await fetch('/webauthn/login/finish', {
        method: 'POST',
        body: JSON.stringify(assertion),
        headers: { 'Content-Type': 'application/json' }
    });
}
```

Security Considerations
-----------------------

1. **Store credentials securely**: Public keys and credential IDs should be
   stored securely in your database.

2. **Verify sign count**: Track and verify the signature counter to detect
   cloned authenticators.

3. **Use attestation carefully**: Attestation can verify authenticator
   properties but has privacy implications.

4. **Fallback authentication**: Consider backup authentication methods for
   users who lose their authenticators.

5. **Multiple credentials**: Allow users to register multiple passkeys
   (phone, laptop, security key).

Migration Strategy
------------------

For existing applications:

1. **Phase 1**: Add passkey support as optional 2FA alongside passwords
2. **Phase 2**: Encourage passkey registration at login
3. **Phase 3**: Allow passwordless login for users with passkeys
4. **Phase 4**: (Optional) Make passkeys the primary authentication method

[1]: https://github.com/go-webauthn/webauthn
