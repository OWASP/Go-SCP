Logging
=======

Logging should always be handled by the application and should not rely on
server configuration.

All logging should be implemented by a master routine on a trusted system, and
the developers should also ensure no sensitive data is included in the logs
(e.g. passwords, session information, system details, etc.), nor is there
any debugging or stack trace information.
Additionally, logging should cover both successful and unsuccessful security
events, with an emphasis on important log event data.

Important event data most commonly refers to:

* All input validation failures.
* All authentication attempts, especially failures.
* All access control failures.
* All apparent tampering events, including unexpected changes to state data.
* All attempts to connect with invalid or expired session tokens.
* All system exceptions.
* All administrative functions, including changes to security configuration
  settings.
* All backend TLS connection failures and cryptographic module failures.

A simple log example which illustrates this:

```go
func main() {
    var buf bytes.Buffer
    var RoleLevel int

    logger := log.New(&buf, "logger: ", log.Lshortfile)

    fmt.Println("Please enter your user level.")
    fmt.Scanf("%d", &RoleLevel) //<--- example

    switch RoleLevel {
    case 1:
        // Log successful login
        logger.Printf("Login successful.")
        fmt.Print(&buf)
    case 2:
        // Log unsuccessful Login
        logger.Printf("Login unsuccessful - Insufficient access level.")
        fmt.Print(&buf)
     default:
        // Unspecified error
        logger.Print("Login error.")
        fmt.Print(&buf)
    }
}
```

It's also good practice to implement generic error messages or custom error
pages as a way to make sure that no information is leaked when an error
occurs.

Go's native package to handle logs doesn't support log levels, which
means that natively to have level based logging would mean implementing
levels by hand. Another issue with the native logger is that there is no
way to turn logging on or off on a per-package basis.

Since all applications require proper logging to maintain upkeep and
security, most of the projects use a third-party logging library like:

* [Logrus][1] - https://github.com/Sirupsen/logrus
* [glog][2]   - https://github.com/golang/glog
* [loggo][3]  - https://github.com/juju/loggo

Of these libraries, the most used is Logrus.

From the log access perspective, only authorized individuals should have
access to the logs.
Developers should also make sure that a mechanism that allows for log
analysis is set in place, as well as guarantee that no untrusted data will
be executed as code in the intended log viewing software or interface.

Regarding allocated memory cleanup, Go has an built-in Garbage Collector
for this very purpose.

As a final step to guarantee log validity and integrity, a cryptographic
hash function should be used as an additional step to ensure no log
tampering has taken place.

```go
{...}
// Get our known Log checksum from checksum file.
logChecksum, err := ioutil.ReadFile("log/checksum")
str := string(logChecksum) // convert content to a 'string'

// Compute our current log's MD5
b, err := ComputeMd5("log/log")
if err != nil {
  fmt.Printf("Err: %v", err)
} else {
  md5Result := hex.EncodeToString(b)
  // Compare our calculated hash with our stored hash
  if str == md5Result {
    // Ok the checksums match.
    fmt.Println("Log integrity OK.")
  } else {
    // The file integrity has been compromised...
    fmt.Println("File Tampering detected.")
  }
}
{...}
```

Note: The `ComputeMD5()` function calculates a file's MD5. It's also important
to note that the log-file hashes must be stored in a safe place, and compared
with the current log hash to verify integrity before any updates to the log.
Full source is included in the document.


[1]: https://github.com/Sirupsen/logrus
[2]: https://github.com/golang/glog
[3]: https://github.com/juju/loggo
