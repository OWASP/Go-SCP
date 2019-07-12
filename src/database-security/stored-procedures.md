Stored Procedures
=================

Developers can use Stored Procedures to create specific views on queries to
prevent sensitive information from being archived, rather than using normal
queries.

By creating and limiting access to stored procedures, the developer is adding
an interface that differentiates who can use a particular stored procedure from
what type of information he can access. Using this, the developer makes the
process even easier to manage, especially when taking control over tables and
columns in a security perspective, which is handy.

Let's take a look into at an example...

Imagine you have a table with information containing users' passport IDs.

Using a query like:

```SQL
SELECT * FROM tblUsers WHERE userId = $user_input
```

Besides the problems of [Input validation][1], the database user (for the
example John) could access __ALL__ information from the user ID.

What if John only has access to use this stored procedure:

```SQL
CREATE PROCEDURE db.getName @userId int = NULL
AS
    SELECT name, lastname FROM tblUsers WHERE userId = @userId
GO
```

Which you can run just by using:

```
EXEC db.getName @userId = 14
```

This way you know for sure that user John only sees `name` and `lastname` from
the users he requests.

Stored procedures are not _bulletproof_, but they create a new layer of
protection to your web application. They give DBAs a big advantage over
controlling permissions (e.g. users can be limited to specific rows/data),
and even better server performance.

[1]: /input-validation/README.md
