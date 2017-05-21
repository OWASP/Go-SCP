Database Connections
====================

## The concept

`sql.Open` does not return a database connection but a handle for database:
_something_ you have available to access database features.
Concerning to connections instead of a single connection, `database/sql`
package manages a pool of connections, what means you have multiple and
concurrent database connections: when you need to perform a database operation,
such as a query, the package takes an available connection from the pool which
should return to the pool as soon as you're "done".
"Done" not always means that your database query was successfully. "done" also
means that the maximum amount of time to complete the operation was exhausted.

You may know that connections (in general file descriptors) are finite so we
should guarantee that we do not loose none of them due to a network hang or
application crash.

The first database connection will be opened only when first required and
`sql.Open` even won't test database connectivity: wrong database credentials
will trigger an error when first database operation runs.

Looking for a _rule of thumb_ a [Context][3] should be always provided and the
context variant of `database/sql` interface (i.e. `QueryContext()` instead of
`Query()`) should be used.

From the official Go documentation "_Package context defines the Context type,
which carries deadlines, cancelation signals, and other request-scoped values
across API boundaries and between processes._".
At a database level when the context is canceled, a transaction will be rolled
back if not committed, a Rows (from QueryContext) will be closed and any
resources will be returned.

```go
package main

import "fmt"
import "context"
import "database/sql"
import "github.com/go-sql-driver/mysql"

func main() {
    ctx := context.Background()
    db, _ := sql.Open("mysql", "user:@/cxdb")
    defer db.Close()

    var version string db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
    fmt.Println("Connected to:", version)
}
```

## Connection string protection

To keep your connection strings secure, it's always a good practice to put the
authentication details on a separated configuration file outside public access.

Instead of placing your configuration file at `/home/public_html/`, consider
`/home/private/configDB.xml` (should be placed in a protected area)

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

Of course, if the attacker has root access, he could see the file. Which brings
us to the most cautious thing you can do - encrypt the file.

## Database Credentials

You should use different credentials for every trust distinction and level:

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
