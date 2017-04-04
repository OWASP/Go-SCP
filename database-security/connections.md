Database Connections
====================

## Keep it closed!

One thing most developers forget is to close the database connection. Let's look
at this example:

```go
package main

import "fmt"
import "database/sql"
import "github.com/go-sql-driver/mysql"

func main() {
    db, _ := sql.Open("mysql", "user:@/cxdb")
    defer db.Close()

    var version string db.QueryRow("SELECT VERSION()").Scan(&version)
    fmt.Println("Connected to:", version)
}
```

As you can see after the opening of the connection, Go shuts down with
[db.Close()][1]. In this case, we used the driver for MariaDB.
Note that the `db.Close()` is inside a `defer` statement. In Go this allows
us to guarantee that the connection is closed.
More information on defer in the [Error Handling and Logging][2] section.

You may ask - *why should I close it?*

After closing the SQL connection, it will be returned to the pool. Furthermore,
since the connections are limited and take resources, if you use the same
connection string, it's possible to reuse it from the pool.

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
