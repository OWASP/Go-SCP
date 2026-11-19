Memory Management
=================

There are several important aspects to consider regarding memory management.
Following the OWASP guidelines, the first step we must take to protect our
application pertains to the user input/output. Steps must be taken to ensure no
malicious content is allowed.
A more detailed overview of this aspect is in the [Input Validation][1] and the
[Output Encoding][2] sections of this document.

Buffer boundary checking is another important aspect of memory management.
checking. When dealing with functions that accept a number of bytes to copy,
usually, in C-style languages, the size of the destination array must be
checked, to ensure we don't write past the allocated space. In Go, data types
such as `String` are not NULL terminated, and in the case of `String`, its
header consists of the following information:

```go
type StringHeader struct {
    Data uintptr
    Len  int
}
```

Despite this, boundary checks must be made (e.g. when looping).
If we go beyond the set boundaries, Go will `Panic`.

Here's a simple example:

```go
func main() {
    strings := []string{"aaa", "bbb", "ccc", "ddd"}
    // Our loop is not checking the MAP length -> BAD
    for i := 0; i < 5; i++ {
        if len(strings[i]) > 0 {
            fmt.Println(strings[i])
        }
    }
}
```

Output:

```
aaa
bbb
ccc
ddd
panic: runtime error: index out of range
```

When our application uses resources, additional checks must also be made to
ensure they have been closed, and not rely solely on the Garbage Collector.
This is applicable when dealing with connection objects, file handles, etc.
In Go we can use `defer` to perform these actions. Instructions in `defer` are
only executed when the surrounding functions finish execution.

```go
defer func() {
    // Our cleanup code here
}
```

More information regarding `defer` can be found in the [Error Handling][3]
section of the document.

Usage of functions that are known to be vulnerable should also be avoided. In
Go, the `Unsafe` package contains these functions. They should not be used in
production environments, nor should the package be used as well. This also
applies to the `Testing` package.

Concurrency and Data Races
--------------------------

In concurrent Go programs, data races are a significant source of security
vulnerabilities and unpredictable behavior. A data race occurs when two or more
goroutines access shared memory concurrently, and at least one access is a write.

Go provides a built-in race detector to find these issues:

```bash
# Run tests with race detection
go test -race ./...

# Build with race detection (for testing only)
go build -race
```

Example of a data race:

```go
// INSECURE - Data race vulnerability
var counter int

func increment() {
    counter++ // Race condition - multiple goroutines modifying shared state
}

func main() {
    for i := 0; i < 1000; i++ {
        go increment()
    }
    // counter value is unpredictable!
}
```

Secure alternatives:

```go
// SECURE - Using sync.Mutex
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}

// SECURE - Using sync/atomic
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

// SECURE - Using channels
func main() {
    counter := 0
    done := make(chan bool)
    increment := make(chan bool)

    go func() {
        for range increment {
            counter++
        }
        done <- true
    }()

    for i := 0; i < 1000; i++ {
        increment <- true
    }
    close(increment)
    <-done
}
```

**Important:** The race detector adds significant overhead (2-10x slower, 5-10x
more memory). Use it during development and testing, not in production.

For comprehensive guidance on security tooling including the race detector,
see the [Security Tooling][4] section.

Memory Deallocation
-------------------

On the other hand, memory deallocation is handled by the garbage collector,
which means that we don't have to worry about it. Please note, it _is_ possible
to manually deallocate memory, although it is _not_ advised.

Quoting [Golang's Github](https://github.com/golang/go/issues/13761):

> If you really want to manually manage memory with Go, implement your own
> memory allocator based on syscall.Mmap or cgo malloc/free.
>
> Disabling GC for extended period of time is generally a bad solution for a
> concurrent language like Go. And Go's GC will only be better down the road.

Memory Leak Scenarios
---------------------

Go is a memory-safe, garbage-collected language, but a program can still keep
memory reachable for much longer than intended. The garbage collector cannot
free anything that is still referenced, even if only a small part of it is
used. Since an attacker can trigger these code paths with crafted input to
exhaust memory and cause a denial of service, it is worth knowing the common
patterns and how to avoid them.

### Substrings that keep a large string alive

Slicing a string does not copy it: the substring shares the backing memory of
the original string. In the function below, the whole of `param` stays in
memory for as long as `prefix` is reachable, even though only 100 bytes of it
are used.

```go
var prefix string // package-level, so it outlives the function call

func keepPrefix(param string) {
    // prefix shares param's backing memory, so all of param stays
    // reachable, not just the first 100 bytes.
    prefix = param[:100]
}
```

Since Go 1.20, the simplest fix is [`strings.Clone`][5], which copies the
substring into new memory so the original string can be garbage collected:

```go
func keepPrefix(param string) {
    prefix = strings.Clone(param[:100])
}
```

On older Go versions, converting to a byte slice and back has the same effect:
`prefix = string([]byte(param[:100]))`.

The same applies to byte slices: `b[:100]` keeps all of `b` alive, while
`bytes.Clone(b[:100])` does not.

### Slices of pointers

When you re-slice a slice of pointers, the elements outside the new bounds are
no longer visible, but they are still stored in the backing array. As long as
the returned slice is in use, the values they point to cannot be garbage
collected.

```go
func middle() []*string {
    items := []*string{new(string), new(string), new(string), new(string), new(string), new(string)}
    // ...
    // items[0] and items[5] are hidden, but the backing array still
    // points to them, so they stay reachable.
    return items[1:5:5]
}
```

To avoid this, set the pointers you no longer need to `nil` before returning
the smaller slice:

```go
func middle() []*string {
    items := []*string{new(string), new(string), new(string), new(string), new(string), new(string)}
    // ...
    // The values items[0] and items[5] point to can now be collected.
    items[0], items[len(items)-1] = nil, nil
    return items[1:5:5]
}
```

The backing array itself is still kept alive. If the original slice is large
and you only need a few elements, copy them into a new slice instead (for
example with `slices.Clone`).

### Deferred calls in loops

Deferred calls run when the surrounding function returns, not at the end of a
loop iteration. In the program below, every file stays open until all files
have been written, so a long list can exhaust file descriptors and memory.

```go
func writeFiles(files []FileDetails) error {
    for _, entry := range files {
        file, err := os.Create(entry.Path)
        if err != nil {
            return err
        }
        // Not run until writeFiles returns: every file stays open
        // until the whole loop has finished.
        defer file.Close()

        if _, err := file.WriteString(entry.Content); err != nil {
            return err
        }
        if err := file.Sync(); err != nil {
            return err
        }
    }

    return nil
}
```

Move the body of the loop into its own function so that each file is closed at
the end of its iteration:

```go
func writeFiles(files []FileDetails) error {
    for _, entry := range files {
        if err := writeFile(entry); err != nil {
            return err
        }
    }

    return nil
}

func writeFile(entry FileDetails) error {
    file, err := os.Create(entry.Path)
    if err != nil {
        return err
    }
    // Runs when writeFile returns, at the end of each loop iteration.
    defer file.Close()

    if _, err := file.WriteString(entry.Content); err != nil {
        return err
    }
    return file.Sync()
}
```

[1]: ../input-validation/README.md
[2]: ../output-encoding/README.md
[3]: ../error-handling-logging/README.md
[4]: ../security-tooling/README.md
[5]: https://pkg.go.dev/strings#Clone
