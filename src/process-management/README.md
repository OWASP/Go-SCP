Process Management
==================
Goroutines is one of the major reasons why Go has been one of the most widely language since past few years.
Goroutines are lightweight thread managed by go runtime which allows developers to write concurrent programs in go.
The cost of goroutines are considerably lighter than a traditional thread and it's very common for go applications to 
have thousands of goroutines concurrently.


However, one of the common problems with goroutines is that a badly written goprogram might produce **hanging** or **zombie** goroutines
which might hold on to the system resources and keep consuming memory. This might cause problems for large-scale services as attacker might
launch attacks which might lead to creation of such hanging Goroutines in large numbers, subsequently leading to putting load on CPU and memory and negatively affecting availablity of the applicaton altogether. Let's look at what hanging goroutines are and ways to avoid those.

Here is an example of what an hanging goroutine looks like:
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

Here even if the handler returns the Goroutine continues live in the background waiting for data to be sent over channel, which will never happen.

There are two common patterns which can result in such leaks. Those patterns are as follows:
1. The Forgotten sender
2. The Abandoned receiver

## The Forgotten Sender
This happens when the sender is blocked because no receiver is waiting on the channel to receive the data.
```go
func danglingSender(ch chan string) {
	val := "Brian"
    // This will be blocked as no receiver is available to receive the data
	ch <- val
}

func handler() {
	ch := make(chan string)

	go danglingSender(ch)
	return
}
```

Such thing can happen in following scenarios.
**Wrong use of Context**

```go
func danglingSender(ch chan string) {
	//simulated async function
	val := someAsyncNetworkCall()
	ch <- val
}

func handlerWithContext() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	ch := make(chan string)
	go danglingSender(ch)

	select {
	case val := <-ch:
		{
			fmt.Printf("Value Received :%s", val)
		}
	case <-ctx.Done():
		{
			return errors.New("Timeout! Exiting")
		}
	}
    return nil
}
	
```
The above program tries to simulate web service handler. We have sent a context that will issue a timeout after 10ms. Then we spawn a Goroutine that will asynchronous network call.  
In case of the network call taking more time than expected and a timeout happens, the program will run run `case <- ctx.Done()` and the handler will return an error.  
As soon as the handler returns the danglingSender function will be blocked as there is no one to receive the val, which will never happen.

**Wrong placement of Receiver**
```go
func danglingSender(ch chan string) {
	val := someAsyncNetworkCall()
	ch <- val
}

func handler() error {
	ch := make(chan string)
	go danglingSender(ch)

	err := validateSomeData()

	if err != nil {
		return errors.New("Validation error.Exiting")
	}

	data := <-ch
	fmt.Println(data)
	return nil
}
```

Here we spawn a Goroutine that makes a network call and simultenously move to some other validation logic. If the validation returns error than the handler exits. And as a consequence the Goroutine gets blocked forever as there is nothing at the receiving end of the channel.


**Solution to The Forgotten Sender**  
To summarize the forgotten sender happens as no receiver is present on the other side to receive the data. And the root cause is the unbuffered channel we have been using.
An unbuffered channel requires a receiver as soon as the message is sent on the channel otherwise the sender is blocked. So, the solution is to buffered channel, where we specifiy the capacity of the channel while initializing it. This way the sender can send data into the channel without requiring a receiver.

```go
func danglingSender(ch chan string) {
	val := "Brian"
    // This won't be blocked
	ch <- val
}

func handler() {
	ch := make(chan string,1)

	go danglingSender(ch)
	return
}
```

## The Abandoned Receiver
This is exact opposite of previous pattern. Here the receiver gets blocked as no sender is sending the data from other side.

```go
func danglingReceiver(ch chan int) {
	//This will block
	val := <-ch
	fmt.Println(val)
}

func handler() {
	ch := make(chan int)
	go danglingReceiver(ch)
}
```

Let's have look at the common scenarios in which this can happen.  
**Sender forgets to close the channel** 

```go
func danglingReceiver(ch chan string) {
	for val := range ch {
		fmt.Println(val)
	}

	fmt.Println("Done.. exiting")
}

func handler(arr []string) {
	ch := make(chan string, len(arr))

	for _, data := range arr {
		ch <- data
	}

	go danglingReceiver(ch)
}
```

Here the Goroutine is expected to process the data sent on the channel by the handler function and terminate. But in reality, the channel even though its empty it's not closed. So, the worker keeps waiting for sender to send more data and never terminates.

**Wrong Placement of Sender**
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
		return errors.New("Validation Error, Exiting..!")
	}

	ch <- data

	return nil
}
```
Here, the handler spawns the Goroutine and quickly moves on to some validation logic. But, in the case of that validation returning an error the handler exits, and there is no one present to send data to the channel. Hence, the receiver is abandoned waiting for the data.


**Solution to The Abandoned Receiver**  
To summarize, the receivers are dangling as they expect incoming data from the channel. Thus, the receiver gets blocked and wait forever.

The solution is to defer closing of the channel.
```go
defer close(ch)
```
It's a recommendation practice to defer the channel's closing as you spawn a new channel.Which will ensure the channel is closed when the function exits.
This can help the receiver to tell if channel is closed and it will terminate accordingly.

```go
func danglingReceiver(ch chan int) {
	//This won't be blocked
	val := <-ch
	fmt.Println(val)
}

func handler() {
	ch := make(chan int)


	//Defer the closing of channel
	defer close(ch)
	go danglingReceiver(ch)
}
```



