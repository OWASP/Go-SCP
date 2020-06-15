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

On the other hand, memory deallocation is handled by the garbage collector,
which means that we don't have to worry about it. Please note, it _is_ possible
to manually deallocate memory, although it is _not_ advised.

Quoting [Golang's Github](https://github.com/golang/go/issues/13761):

> If you really want to manually manage memory with Go, implement your own
> memory allocator based on syscall.Mmap or cgo malloc/free.
>
> Disabling GC for extended period of time is generally a bad solution for a
> concurrent language like Go. And Go's GC will only be better down the road.

[1]: ../input-validation/README.md
[2]: ../output-encoding/README.md
[3]: ../error-handling-logging/README.md
