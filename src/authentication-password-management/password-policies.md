Password Policies
=================

Passwords are a historical asset, part of most authentication systems, and are
the number one target of attackers.

Quite often some service leaks its users' database, and despite the leak of
email addresses and other personal data, the biggest concern are passwords. Why?
Because passwords are not easy to manage and remember. Users not only tend to
use weak passwords (e.g. "123456") they can easily remember, they can also
re-use the same password for different services.

Modern Password Guidelines (NIST 800-63B)
-----------------------------------------

Password guidance has evolved significantly. [NIST Special Publication 800-63B][1]
provides updated recommendations that differ from historical practices:

**Length over complexity:**
- Require a minimum of 8 characters (NIST minimum)
- Allow at least 64 characters maximum
- Consider multi-word passphrases (e.g., "correct-horse-battery-staple")
- **Do NOT** require arbitrary complexity rules (uppercase, symbols, etc.) as
  they lead to predictable patterns

**Password checks:**
- Check passwords against known breach databases (e.g., [Have I Been Pwned][2])
- Block common passwords (e.g., "password", "123456", company name)
- Block context-specific passwords (username, email, service name)

**What NOT to do (contrary to legacy advice):**
- **Do NOT** require periodic password rotation - This leads to weaker passwords
  as users make predictable changes (Password1 → Password2)
- **Do NOT** use password hints or knowledge-based authentication (security
  questions) - These are vulnerable to social engineering and data breaches
- **Do NOT** require SMS-based 2FA for sensitive applications - Use
  authenticator apps or hardware tokens instead

**Encourage MFA:**
The best way to protect accounts is Multi-Factor Authentication. Consider
supporting:
- TOTP authenticator apps (Google Authenticator, Authy)
- Hardware security keys (WebAuthn/FIDO2)
- Passkeys (see [Passkeys section](passkeys.md))

Implementation Example
----------------------

```go
// Check password against common passwords and breach databases
func validatePassword(password, username, email string) error {
    // Minimum length
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }

    // Check against context-specific strings
    lowerPass := strings.ToLower(password)
    if strings.Contains(lowerPass, strings.ToLower(username)) {
        return errors.New("password cannot contain username")
    }

    // Check against common passwords (maintain a blocklist)
    if isCommonPassword(password) {
        return errors.New("password is too common")
    }

    // Optionally check against breach databases
    if isBreachedPassword(password) {
        return errors.New("password has appeared in a data breach")
    }

    return nil
}
```

## Reset

Even if you're not applying any extra password policy, users still need to be
able to reset their password.
Such a mechanism is as critical as signup or sign-in, and you're encouraged to
follow the best practices to be sure your system does not disclose sensitive
data and become compromised.

**Secure Password Reset Flow:**

1. **Use email-based resets only** - Send a reset link (not a temporary password)
   to a pre-registered email address
2. **Use cryptographically secure tokens** - Generate random tokens using
   `crypto/rand`, not predictable values
3. **Set short expiration** - Reset links should expire within 1 hour maximum
4. **Single use** - Invalidate the token after use or after a new one is generated
5. **Notify the user** - Send an email notification when a password is changed
6. **Rate limit requests** - Prevent enumeration and abuse

```go
// Generate a secure reset token
func generateResetToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

**Avoid these legacy practices:**
- **Security questions** - Answers are often guessable or found on social media
- **Sending passwords via email** - Always use one-time reset links instead
- **SMS-based reset** - Vulnerable to SIM swapping attacks

[1]: https://pages.nist.gov/800-63-3/sp800-63b.html
[2]: https://haveibeenpwned.com/API/v3
