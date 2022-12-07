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
[3]: ../error-handling-logging/README.

## Memory Leaking Scenarios
Even though go is a memory-safe language. The go compiler has been written in a way that can cause it to kind-of memory leaking issues sometimes. Since memory leaking is one of the ways using which an attacker can launch attacks like DDoS, one needs to be aware of recommended practices to prevent writing programs that can lead to memory leaking.Let's look at some of the scenarios.

## Memory leak. because of substrings:
For example, in the below function the memory occupied by the paramStr variable won't be garbage collected even though the function has returned as the strSlice variable is sharing the same underlying memory block.

```go
var strSlice string //variable of type string on heap

func customizedSubstring(paramStr string) {
	strSlice = paramStr[:100]
	/*
		Since strSlice is created by slicing the paramStr, both the variables
		share the same underlying memory block. Even though paramStr's lifespan
		finishes after the function has returned, that memory can't be garbage
		collected as some chunk of it's memory(100 bytes) is in use.
	*/
}

```
To get around such scenario, we can opt for multiple ways to efficiently copy a string, few of those are listed below:
## Using a byte array

```go
func customizedSubstring(paramStr string) {
	tempByteArray := []byte(paramStr[:100])
	strSlice = string(tempByteArray)
	/*
		A two step process where we first convert the string to a byte array whic
		then convert that intermediate byte array to a string again. This way the
		resulting slice gets created at new memory block.
	*/
}
```

# Concatenation
```go
func customizedSubstring(paramStr string) {
	strSlice = (" " + paramStr[:100][:1])
	/*
		The use of empty string makes the compiler creates a new memory block for
		resulting string at the cost of 1 extra byte.
	*/
}
```

# Using a string Builder
```go
func customizedSubstring(paramStr string) {
	var temp strings.Builder
	temp.Grow(100)
	temp.WriteString(paramStr[:100])
	strSlice = temp.String()

	/*
		A slightly verbose way of making sure that sliced string gets created at a new memory
		location and the source string can be collected by the garbage collector.
	*/
}
```

## Memory Leaking with Pointers:
Let's look at the function below. Once the function's lifespan is over, the memory reference allocated for the
the first and the last elements of the slice will be lost.

```go
func slicePointerFunc() []*string {
	new_slice := []*string{new(string), new(string), new(string), new(string), new(string), new(string)}
	//.....
	return new_slice[1:4:4]
	/*As long as this returned slice is been in use in the other parts of the program
	the underling slice can't be garbage collected. Subsequently preventing first and
	last elements of the slice from  being garbage collected*/
}
```

To avoid such memory leaking, we can reset the unused pointers by simply setting them to nil.

```go
func slicePointerFunc() []*string {
	new_slice := []*string{new(string), new(string), new(string), new(string), new(string), new(string)}
	//....
	new_slice[0], new_slice[len(new_slice)-1] = nil, nil
	/*Marking the unused elements of slice as nil, will make them available for garbage collection */
	return new_slice[1:4:4]
}
```

## Memory leaking because of deferred function calls:
Sometimes a deferred call queue can consume significant memory and may hold on to resources which are needed by other programs/subprograms running on the system. For example, look at the below program, which writes to number of files during it's lifespan.
Such program might hold on to the file handlers till the very end of the function, even after finishing updating those files much earlier.
```go
func multiFileWrites(fileList []FileDetails) error {
	for _, entry := range fileList {
		file, err := os.Open(entry.Path)
		if err != nil {
			return err
		}

		defer file.Close()
		/* Such approach may result in large number of deferred calls in case
		of thousands of file writes. Subsequently delaying the release of system
		resources.
		*/
		_, err = file.WriteString(entry.Content)
		if err != nil {
			return err
		}

		err = file.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}
```

So, the solution in such cases is to use an anonymous function that will enclose the deferred calls to make them execute relatively earlier.

```go
func memoryEfficientMultipleFileWrites(fileList []FileDetails) error {
	for _, entry := range fileList {
		if err := func() error {
			file, err := os.Open(entry.Path)
			if err != nil {
				return err
			}

			/*
				Use of anonymous function makes sure that the every deferred call
				for file close gets called at the end of every iteration instead of
				processing every call at the end of the outer function.
			*/
			defer file.Close()

			_, err = file.WriteString(entry.Content)
			if err != nil {
				return err
			}

			return file.Sync()
		}(); err != nil {
			return err
		}

	}

	return nil
}
```
