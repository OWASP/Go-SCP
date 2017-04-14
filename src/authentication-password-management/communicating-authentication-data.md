Communicating authentication data
=================================

In this section, "communication" is used in a broader sense, encompassing
User Experience (UX) and client-server communication.

Not only is it true that "_password entry should be obscured on user's screen_"
but also the "_remeber me functionality should be disabled_".

You can accomplish both using an input field with `type="password"`, and
setting the `autocomplete` attribute to `off`[^1]

```html
<input type="password" name="passwd" autocomplete="off" />
```

Authentication credentials should be sent on HTTP POST requests only, using an
encrypted connection (HTTPS). An exception to the encrypted connection may be
the temporary passwords associated with email resets.

Although HTTP GET requests over TLS/SSL (HTTPS) look as secure as HTTP POST
requests, remember that in general HTTP servers (eg. Apache[^2], Nginx[^3]) do
write the requested URL to the access log.

```text
xxx.xxx.xxx.xxx - - [27/Feb/2017:01:55:09 +0000] "GET /?username=user&password=70pS3cure/oassw0rd HTTP/1.1" 200 235 "-" "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:51.0) Gecko/20100101 Firefox/51.0"
```

A well designed HTML form for authentication would look like:

```html
<form method="post" action="https://somedomain.com/user/signin" autocomplete="off">
    <input type="hidden" name="csrf" value="CSRF-TOKEN" />

    <label>Username <input type="text" name="username" /></label>
    <label>Password <input type="password" name="password" /></label>

    <input type="submit" value="Submit" />
</form>
```

When handling authentication errors, your application should not disclose which
part of the authentication data was incorrect. Instead of "Invalid username" or
"Invalid password", just use "Invalid username and/or password" interchangeably:

```html
<form method="post" action="https://somedomain.com/user/signin" autocomplete="off">
    <input type="hidden" name="csrf" value="CSRF-TOKEN" />

    <div class="error">
        <p>Invalid username and/or password</p>
    </div>

    <label>Username <input type="text" name="username" /></label>
    <label>Password <input type="password" name="password" /></label>

    <input type="submit" value="Submit" />
</form>
```

With a generic message you do not disclose:

* Who is registered: "Invalid password" means that the username exists.
* How your system works: "Invalid password" reveals how your application works

```go
  record, err = db.Query("SELECT password FROM accounts WHERE username = ?", username)
  if err != nil {
    // user does not exist
  }

  if record[0] != password {
    // passwords do not match
  }
```

After a successful login, the user should be informed about the last successful
or unsuccessful access date/time so that he can detect and report suspicious
activity. Further information regarding logging can be found in the
[`Error Handling and Logging`][4] section of the document.

---

[^1]: [How to Turn Off Form Autocompletion][1], Mozilla Developer Network
[^2]: [Log Files][2], Apache Documentation
[^3]: [log_format][3], Nginx log_module "log_format" directive

[1]: https://developer.mozilla.org/en-US/docs/Web/Security/Securing_your_site/Turning_off_form_autocompletion
[2]: https://httpd.apache.org/docs/1.3/logs.html#accesslog
[3]: http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
[4]: ../error-handling-logging/logging.md
