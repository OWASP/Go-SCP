Parameterized Queries
=====================

Prepared Statements (with Parameterized Queries) are the best and most secure
way to protect against SQL Injections.

In some reported situations, prepared statements could harm performance of the
web application. Therefore, if for any reason you need to stop using this type
of database queries, we strongly suggest you read [Input Validation][1] and
[Output Encoding][2] sections.

Go works differently from usual prepared statements on other languages - you
don't prepare a statement on a connection. You prepare it on the DB.

## Flow

1. The developer prepares the statement (`Stmt`) on a connection in the pool
2. The `Stmt` object remembers which connection was used
3. When the application executes the `Stmt`, it tries to use that connection.
   If it's not available it will try to find another connection in the pool

This type of flow could cause high-concurrency usage of the database and creates
lots of prepared statements. Therefore, it's important to keep this information
in mind.

Here's an example of a prepared statement with parameterized queries:

```go
customerName := r.URL.Query().Get("name")
db.Exec("UPDATE creditcards SET name=? WHERE customerId=?", customerName, 233, 90)
```

Sometimes a prepared statement is not what you want. There might be several
reasons for this:

* The database doesn’t support prepared statements. When using the MySQL driver,
  for example, you can connect to MemSQL and Sphinx, because they support the
  MySQL wire protocol. But they don’t support the "binary" protocol that
  includes prepared statements, so they can fail in confusing ways.

* The statements aren’t reused enough to make them worthwhile, and security
  issues are handled in another layer of our application stack
  (See: [Input Validation][1] and [Output Encoding][2]), so performance
  as seen above is undesired.

[1]: ../input-validation/README.md
[2]: ../output-encoding/README.md
