Database Connections
====================

## The concept

`sql.Open` does not return a database connection but `*DB`: a database
connection pool. When a database operation is about to run (e.g. query), an
available connection is taken from the pool, which should be returned to the
pool as soon as the operation completes.

Remember that a database connection will be opened only when first required to
perform a database operation, such as a query.
`sql.Open` doesn't even test database connectivity: wrong database credentials
will trigger an error at the first database operation execution time.

Looking for a _rule of thumb_, the context variant of `database/sql` interface
(e.g. `QueryContext()`) should always be used and provided with the appropriate
[Context][3].

From the official Go documentation "_Package context defines the Context type,
which carries deadlines, cancelation signals, and other request-scoped values
across API boundaries and between processes._".
At a database level, when the context is canceled, a transaction will be rolled
back if not committed, a Rows (from QueryContext) will be closed and any
resources will be returned.

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

type program struct {
    base context.Context
    cancel func()
    db *sql.DB
}

func main() {
    db, err := sql.Open("mysql", "user:@/cxdb")
    if err != nil {
        log.Fatal(err)
    }
    p := &program{db: db}
    p.base, p.cancel = context.WithCancel(context.Background())

    // Wait for program termination request, cancel base context on request.
    go func() {
        osSignal := // ...
        select {
        case <-p.base.Done():
        case <-osSignal:
            p.cancel()
        }
        // Optionally wait for N milliseconds before calling os.Exit.
    }()

    err =  p.doOperation()
    if err != nil {
        log.Fatal(err)
    }
}

func (p *program) doOperation() error {
    ctx, cancel := context.WithTimeout(p.base, 10 * time.Second)
    defer cancel()

    var version string
    err := p.db.QueryRowContext(ctx, "SELECT VERSION();").Scan(&version)
    if err != nil {
        return fmt.Errorf("unable to read version %v", err)
    }
    fmt.Println("Connected to:", version)
}
```

## Connection string protection

To keep your connection strings secure, it's always a good practice to put the
authentication details on a separated configuration file, outside of public
access.

Instead of placing your configuration file at `/home/public_html/`, consider
`/home/private/configDB.xml`: a protected area.

```xml
<connectionDB>
  <serverDB>localhost</serverDB>
  <userDB>f00</userDB>
  <passDB>f00?bar#ItsP0ssible</passDB>
</connectionDB>
```

Then you can call the configDB.xml file on your Go file:

```go
configFile, _ := os.Open("../private/configDB.xml")
```

After reading the file, make the database connection:

```go
db, _ := sql.Open(serverDB, userDB, passDB)
```

Of course, if the attacker has root access, he will be able to see the file.
Which brings us to the most cautious thing you can do - encrypt the file.

## Database Credentials

You should use different credentials for every trust distinction and level, for
example:

* User
* Read-only user
* Guest
* Admin

That way if a connection is being made for a read-only user, they could never
mess up with your database information because the user actually can only read
the data.

[1]: https://golang.org/pkg/database/sql/#DB.Close
[2]: ../error-handling-logging/README.md
[3]: https://golang.org/pkg/context/
