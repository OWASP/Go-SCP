SQL Injection
=============

Another common injection due to the lack of proper output encoding is SQL
Injection, mostly because of an old bad practice: string concatenation.

In short: whenever a variable holding a value which may include arbitrary
characters such as ones with special meaning to the database management system
is simply added to a (partial) SQL query, you're vulnerable to SQL Injection.

Imagine you have a query such as the one below:

```go
customerId := r.URL.Query().Get("id")
query := "SELECT number, expireDate, cvv FROM creditcards WHERE customerId = " + customerId

row, _ := db.Query(query)
```
You’re about to ruin your life.

When provided a valid `customerId` you will list only that customer's credit
cards, but what if `customerId` becomes `1 OR 1=1`?

Your query will look like:

```SQL
SELECT number, expireDate, cvv FROM creditcards WHERE customerId = 1 OR 1=1
```

... and you will dump all table records (yes, `1=1` will be true for any record)!

There's only one way to keep your database safe: [Prepared Statements][1].

```go
customerId := r.URL.Query().Get("id")
query := "SELECT number, expireDate, cvv FROM creditcards WHERE customerId = ?"

stmt, _ := db.Query(query, customerId)
```
Notice the placeholder `?` and how your query is:

 * readable,
 * shorter and
 * SAFE

Placeholder syntax in prepared statements is database-specific.
For example, comparing MySQL, PostgreSQL, and Oracle:

| MySQL | PostgreSQL | Oracle |
| :---: | :--------: | :----: |
| WHERE col = ? | WHERE col = $1 | WHERE col = :col |
| VALUES(?, ?, ?) | VALUES($1, $2, $3) | VALUES(:val1, :val2, :val3) |

Check Database Security section in this guide to get more in-depth information
about this topic.

[1]: https://golang.org/pkg/database/sql/#DB.Prepare
