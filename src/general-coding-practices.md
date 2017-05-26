General Coding Practices
========================

There are a few general guidelines you should consider while developing
software.

* "_Use tested and approved managed code rather than creating new unmanaged
  code for common tasks_"

  Quite often we see exactly the same mistakes, bugs and/or vulnerabilities.
  One of the common causes for that is the fact that we're used to approach
  problems with "vanilla code": code written from scratch, not tested or
  maintained.
  Whenever possible, opt for managed code such as frameworks; as they are
  developed, tested and used by many people, issues would arise and get fixed
  earlier.

* "_Utilize task specific built-in APIs to conduct operating system tasks. Do
  not allow the application to issue commands directly to the Operating System,
  especially through the use of application initiated command shells_"

  Almost all programming languages allow you to initiate a command shell as Go
  does.

  ```go
  // Cat (command) a file example
  // set FS permissions to a given (by the user) file
  func main() {
      reader := bufio.NewReader(os.Stdin)
      // Ask the user what file to be read
      file, _ := reader.ReadString('\n')
      if err := exec.Command("cat", "-A", file).Run(); err != nil {
          fmt.Fprintln(os.Stderr, err)
          os.Exit(1)
      }
      fmt.Print("Executed command -> ")
      fmt.Println(file)
      fmt.Println("Command successful.")
  }
  ```

  At first, it would look like a nice way to perform low level tasks, but you're
  just creating a security breach if you're not careful and, for example call
  the OS shell directly with the `-c` argument.
  Using `exec.Command()` is safe as long as it's not executing a binary that
  accepts a program as an argument as demonstrated here with the `bash` and
  `-c` command.

  ```go
  // pass file name as 'file.png; rm -rf / #
  if err := exec.Command("bash", "-c", input).Run(); err != nil {
      fmt.Fprintln(os.Stderr, err)
      os.Exit(1)
  }
  ```

  Always use task specific built-in APIs

  ```go
  if err := os.Chmod(file, 0644); err != nil {
      log.Fatal(err)
  }
  ```

* "_Use checksums or hashes to verify the integrity of interpreted code,
  libraries, executables, and configuration files_"

  If your application relies on third party resources such as libraries or
  configuration files, how can you be sure that at execution time they remain
  exactly as they were when they were deployed?

  Or even worse, if your application loads third party scripts from remote
  hosts, what kind of warranty do you have that the file won't change, thus
  breaking your application?

  Maybe you're thinking about CDNs - Content Delivery Networks. They are
  everywhere and we "need" them. But what if they get compromised and resources
  get modified somehow?

  Have a look on the [Subresource Integrity][1] section.
  How could did we live without it for such a long time!?

* "_Utilize locking to prevent multiple simultaneous requests or use a
  synchronization mechanism to prevent race conditions_"

  Race condition is what you have when a shared resource gets accessed
  simultaneously by multiple requesters. Who gets the right to access the
  shared resource?

  This is an old problem, quite common in concurrent environments.
  The solution is also often enough not taken into account.

  The best approach to this is to use Mutexes which are available in Go's
  `sync` package. A simple example taken from the "Go Tour":

  ```go
  package main

  import (
  	"fmt"
  	"sync"
  	"time"
  )

  // SafeCounter is safe to use concurrently.
  type SafeCounter struct {
  	v   map[string]int
  	mux sync.Mutex
  }

  // Inc increments the counter for the given key.
  func (c *SafeCounter) Inc(key string) {
  	c.mux.Lock()
  	// Lock so only one goroutine at a time can access the map c.v.
  	c.v[key]++
  	c.mux.Unlock()
  }

  // Value returns the current value of the counter for the given key.
  func (c *SafeCounter) Value(key string) int {
  	c.mux.Lock()
  	// Lock so only one goroutine at a time can access the map c.v.
  	defer c.mux.Unlock()
  	return c.v[key]
  }

  func main() {
  	c := SafeCounter{v: make(map[string]int)}
  	for i := 0; i < 1000; i++ {
  		go c.Inc("somekey")
  	}

  	time.Sleep(time.Second)
  	fmt.Println(c.Value("somekey"))
  }
  ```

  Another problem is resource exhaustion, which can lead to Denial of Service.  
  Although there is no native support for semaphores in Go, they can be recreated
  using buffered channels.

  A few examples of the usage of semaphores:

      - Database connections
      - TCP/IP output connections
      - Threads
      - Memory

  A simple example of semaphore usage in Go:

  ```go
  // write to file
  const (
      AvailableMemory         = 10 << 20 // 10 MB
      AverageMemoryPerRequest = 10 << 10 // 10 KB
      MaxOutstanding          = AvailableMemory / AverageMemoryPerRequest
  )

  var sem = make(chan int, MaxOutstanding)

  func Serve(queue chan *Request) {
      for {
          sem <- 1 // Block until there's capacity to process a request.
          req := <-queue
          go handle(req) // Don't wait for handle to finish.
      }
  }

  func handle(r *Request) {
      process(r) // May take a long time & use a lot of memory or CPU
      <-sem      // Done; enable next request to run.
  }
  ```

* "_Protect shared variables and resources from inappropriate concurrent
  access_"

  By now you already know how to approach this problem; using a mutex or a
  semaphore would solve any further issues.

* "_Explicitly initialize all your variables and other data stores, either
  during declaration or just before the first usage_"


* "_In cases where the application must run with elevated privileges, raise
  privileges as late as possible, and drop them as soon as possible_"


* "_Avoid calculation errors by understanding your programming language's
  underlying representation and how it interacts with numeric calculation. Pay
  close attention to byte size discrepancies, precision, signed/unsigned
  distinctions, truncation, conversion and casting between types, "not-a-number"
  calculations, and how your language handles numbers that are too large or too
  small for its underlying representation_"

  You should always remember that even the best programming language will have
  to deal with hardware limitations. One limitation we usually tend to forget
  is the floating number representation lack of precision.

  ```go
  package main

  import "fmt"

  func main () {
      var n float64 = 0

      for i := 0; i < 10; i++ {
          n += .1
      }

      fmt.Println(n)
  }
  ```

  You may expect the result of summing `0.1` ten times to be `1` but what
  you'll get is:

  ```bash
  0.9999999999999999
  ```

  See what happens when dealing with large numbers:

  ```go
  package main

  import "fmt"
  import "math"

  func main () {
      var n int64 = math.MaxInt64

      fmt.Println(n)
      fmt.Println(n + 1)
  }
  ```

  ```bash
  9223372036854775807
  -9223372036854775808
  ```

  All you need is a library to handle big numbers: [math/big package][4]

  ```go
  package main

  import "fmt"
  import "math"
  import "math/big"

  func main () {
      n1 := new(big.Int).SetInt64(math.MaxInt64)
      n2 := new(big.Int).SetInt64(1)
      sum := new(big.Int)

      fmt.Println(n1)
      fmt.Println(sum.Add(n1, n2))
  }
  ```

  And, as expected, you'll get:

  ```bash
  9223372036854775807
  9223372036854775808
  ```

* "_Do not pass user supplied data to any dynamic execution function_"

  For more information, continue reading the [Input Validation][2] and
  [Output Encoding][3] sections; there's no shortcut to take.

* "_Restrict users from generating new code or altering existing code_"

  There are a few use cases in which users are supposed to upload source code
  to run server side. If you have a need for this, you should do
  it in a restricted environment, otherwise you will lose control.

  Let's start from the beginning.

  Your application's source code files should not be writable, making them read
  only or, at most, executable. This will prevent an attacker who's able to
  exploit your application by adding extra source code, getting it to run and
  maybe open a shell to gain control over your server.

  In the same way, uploaded files permissions should be set accordingly. Usually
  read-only will be just fine; pictures/photos, spreadsheets, text documents,
  etc... They won't need execution permission.

  When dealing with image files, you should pre-process them server side,
  converting them to a safe and standard format, avoiding script injection
  through image files metadata. Images with EXIF tag processing should follow
  the [Output Encoding][3] guidelines as they may contain malicious code.

  ```go
  // Open out file to be converted
  imageFile, err := os.Open("logo.jpg")
  if err != nil {
      fmt.Println("Error opening file.")
  }

  // decode jpeg into image.Image
  imageDecoded, err := jpeg.Decode(imageFile)

  // Create the new image file
  out, err := os.Create("logo.png")

  // Encode the image to png
  err = png.Encode(out, imageDecoded)
  ```

  If for any special reason you have to evaluate user input as source
  code, do it only in a sandboxed environment[^1].

* "_Review all secondary applications, third party code and libraries to
  determine business necessity and validate safe functionality, as these can
  introduce new vulnerabilities_"

  You should audit every single third party library added to your project,
  they will be a part your application, running with the same access rights
  and/or privileges.

* "_Implement safe updating. If the application will utilize automatic updates,
  then use cryptographic signatures for your code and ensure your download
  clients verify those signatures. Use encrypted channels to transfer the code
  from the host server_"

---

[^1]: ["Inside the Go Playground"][5], The Go Blog, December 2013


[1]: https://www.w3.org/TR/SRI/
[2]: /input-validation
[3]: /output-encoding
[4]: https://golang.org/pkg/math/big/
[5]: https://blog.golang.org/playground
