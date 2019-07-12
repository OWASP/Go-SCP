SQL Injection
=============

Another common injection that's due to the lack of proper output encoding is SQL
Injection. This is mostly due to an old bad practice: string concatenation.

In short, whenever a variable holding a value which may include arbitrary
characters such as ones with special meaning to the database management system
is simply added to a (partial) SQL query, you're vulnerable to SQL Injection.

Imagine you have a query such as the one below:

```go
ctx := context.Background()
customerId := r.URL.Query().Get("id")
query := "SELECT number, expireDate, cvv FROM creditcards WHERE customerId = " + customerId

row, _ := db.QueryContext(ctx, query)
```

You're about to be exploited and subsequently breached.

For example, when provided a valid `customerId` value you will only list that
customer's credit card(s). But what if `customerId` becomes `1 OR 1=1`?

Your query will look like:

```SQL
SELECT number, expireDate, cvv FROM creditcards WHERE customerId = 1 OR 1=1
```

... and you will dump all table records (yes, `1=1` will be true for any record)!

There's only one way to keep your database safe: [Prepared Statements][1].

```go
ctx := context.Background()
customerId := r.URL.Query().Get("id")
query := "SELECT number, expireDate, cvv FROM creditcards WHERE customerId = ?"

stmt, _ := db.QueryContext(ctx, query, customerId)
```
Notice the placeholder `?`. Your query is now:

 * readable,
 * shorter and
 * SAFE

Placeholder syntax in prepared statements is database-specific.
For example, comparing MySQL, PostgreSQL, and Oracle:

| MySQL | PostgreSQL | Oracle |
| :---: | :--------: | :----: |
| WHERE col = ? | WHERE col = $1 | WHERE col = :col |
| VALUES(?, ?, ?) | VALUES($1, $2, $3) | VALUES(:val1, :val2, :val3) |

Check the Database Security section in this guide to get more in-depth
information about this topic.

[1]: https://golang.org/pkg/database/sql/#DB.Prepare
