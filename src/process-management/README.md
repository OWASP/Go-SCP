Process Management
==================

Goroutines are one of the main reasons for Go's popularity. They are
lightweight threads managed by the Go runtime, and they make it easy to write
concurrent programs. A goroutine costs far less than an operating system
thread, so it is common for Go applications to run thousands of them at once.

However, a badly written program can leave **leaked** goroutines behind:
goroutines that are blocked forever and will never exit. Each one keeps its
stack and everything it references in memory. If an attacker can trigger the
code path that leaks a goroutine, for example by sending requests that time
out, they can make the service accumulate leaked goroutines until it runs out
of memory, which hurts the availability of the application. Let's look at what
leaked goroutines look like and how to avoid them.

Here is an example of a leaked goroutine:

```go
func leakingGoroutine(ch chan string) {
    val := <-ch
    fmt.Println(val)
}

func handler() {
    ch := make(chan string)

    go leakingGoroutine(ch)
    return
}
```

Even after `handler` returns, the goroutine stays alive in the background,
waiting for a value on the channel that will never be sent.

There are two common patterns that cause such leaks:

1. The forgotten sender
2. The abandoned receiver

## The Forgotten Sender

This happens when a sender is blocked because no receiver will ever read from
the channel.

```go
func danglingSender(ch chan string) {
    val := "Brian"
    // This blocks forever, as no receiver is available.
    ch <- val
}

func handler() {
    ch := make(chan string)

    go danglingSender(ch)
    return
}
```

This can happen in the following scenarios.

**Returning early after a timeout**

```go
func danglingSender(ch chan string) {
    val := someNetworkCall() // may take longer than the caller waits
    ch <- val
}

func handlerWithContext() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
    defer cancel()

    ch := make(chan string)
    go danglingSender(ch)

    select {
    case val := <-ch:
        fmt.Printf("Value received: %s\n", val)
    case <-ctx.Done():
        return errors.New("timeout")
    }
    return nil
}
```

This simulates a web service handler. The context times out after 10ms while
a goroutine makes a network call. If the call takes longer than that, the
handler takes the `<-ctx.Done()` branch and returns an error. When the network
call finally finishes, `danglingSender` blocks forever trying to send `val`,
because nobody is receiving anymore.

**Returning early before receiving**

```go
func danglingSender(ch chan string) {
    val := someNetworkCall()
    ch <- val
}

func handler() error {
    ch := make(chan string)
    go danglingSender(ch)

    if err := validateSomeData(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    data := <-ch
    fmt.Println(data)
    return nil
}
```

Here we start a goroutine that makes a network call and, at the same time, run
some validation. If the validation fails, the handler returns, and the goroutine
blocks forever because nothing will receive from the channel.

**Solution to the forgotten sender**

The forgotten sender happens because nothing receives the value, and the
channel is unbuffered: a send on an unbuffered channel blocks until a receiver
takes the value. The solution is to give the channel enough buffer space for
every value the goroutine will send, so the send completes even if nobody ever
reads it:

```go
func danglingSender(ch chan string) {
    val := "Brian"
    // This no longer blocks: the value goes into the buffer.
    ch <- val
}

func handler() {
    ch := make(chan string, 1)

    go danglingSender(ch)
    return
}
```

The same fix (`make(chan string, 1)`) applies to both scenarios above. Once the
goroutine has sent its value and returned, the channel and the value are
garbage collected.

## The Abandoned Receiver

This is the opposite pattern: a receiver is blocked because no sender will ever
send on the channel or close it.

```go
func danglingReceiver(ch chan int) {
    // This blocks forever.
    val := <-ch
    fmt.Println(val)
}

func handler() {
    ch := make(chan int)
    go danglingReceiver(ch)
}
```

Let's look at the common scenarios in which this happens.

**The sender forgets to close the channel**

```go
func danglingReceiver(ch chan string) {
    for val := range ch {
        fmt.Println(val)
    }

    fmt.Println("Done, exiting")
}

func handler(arr []string) {
    ch := make(chan string, len(arr))

    for _, data := range arr {
        ch <- data
    }

    go danglingReceiver(ch)
}
```

The goroutine is supposed to process the values the handler sent and then
exit. But although the channel is empty after the values are read, it is never
closed, so the `range` loop keeps waiting for more values and the goroutine
never exits. Calling `close(ch)` after the last send fixes it.

**Returning early before sending**

```go
func danglingReceiver(ch chan []string) {
    val := <-ch
    fmt.Println(val)
}

func handler() error {
    ch := make(chan []string)
    go danglingReceiver(ch)

    data, err := someOtherValidationLogic()
    if err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    ch <- data
    return nil
}
```

The handler starts the goroutine and moves on to some validation. If the
validation fails, the handler returns and nothing ever sends on the channel, so
the receiver waits forever.

**Solution to the abandoned receiver**

The receiver is stuck because it waits for a value that never comes. The
solution is for the sending side to close the channel when it is done, so the
receiver can see that no more values will arrive. Deferring the close right
after creating the channel guarantees it happens on every return path,
including early returns:

```go
func danglingReceiver(ch chan int) {
    // Receives the zero value once ch is closed, so this doesn't block forever.
    val, ok := <-ch
    if !ok {
        return
    }
    fmt.Println(val)
}

func handler() {
    ch := make(chan int)
    // Close the channel on every return path.
    defer close(ch)

    go danglingReceiver(ch)
}
```

Only the sender should close a channel, and only once: sending on a closed
channel, or closing it twice, panics.

## Preventing and Detecting Leaks

* Give every goroutine a clear way to exit, such as a closed channel or a
  cancelled [context][1]. See the [Context Timeouts][2] section for patterns
  that tie goroutines to the lifetime of a request.
* Watch the number of running goroutines in production, for example with
  `runtime.NumGoroutine()` or the `/debug/pprof/goroutine` endpoint of
  [net/http/pprof][3] (which should never be exposed publicly).
* Check for leaks in tests with a package such as [goleak][4], which fails a
  test if goroutines are still running when it finishes.

[1]: https://pkg.go.dev/context
[2]: ../error-handling-logging/context-timeouts.md
[3]: https://pkg.go.dev/net/http/pprof
[4]: https://github.com/uber-go/goleak
