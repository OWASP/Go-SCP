Database Security
=================

This section on OWASP SCP will cover all of the database security issues and
actions developers and DBAs need to take when using databases in their web
applications.

Go doesn't have database drivers. Instead there is a core interface driver on
the [database/sql][1] package. This means that you need to register your SQL
driver (eg: [MariaDB][2], [sqlite3][3]) when using database connections.

## The best practice

Before implementing your database in Go, you should take care of some
configurations that we'll cover next:

* Secure database server installation[^1].
    * Change/set a password for `root` account(s).
    * Remove the `root` accounts that are accessible from outside the localhost.
    * Remove any anonymous-user accounts.
    * Remove any existing test database.
* Remove any unnecessary stored procedures, utility packages,
  unnecessary services, vendor content (e.g. sample schemas).
* Install the minimum set of features and options required for your database to
  work with Go.
* Disable any default accounts that are not required on your web application to
  connect to the database.

Also, because it's __important__ to validate input, and encode output on the
database, be sure to investigate the [Input Validation][4] and [Output
Encoding][5] sections of this guide.

This basically can be adapted to any programming language when using databases.

---

[^1]: MySQL/MariaDB have a program for this: `mysql_secure_installation`<sup>[1][6], [2][7]</sup>

[1]: https://golang.org/pkg/database/sql/
[2]: https://github.com/go-sql-driver/mysql
[3]: https://github.com/mattn/go-sqlite3
[4]: ../input-validation/README.md
[5]: ../output-encoding/README.md
[6]: https://dev.mysql.com/doc/refman/5.7/en/mysql-secure-installation.html
[7]: https://mariadb.com/kb/en/mariadb/mysql_secure_installation/
