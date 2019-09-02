Database Authentication
=======================

## Access the database with minimal privilege

If your Go web application only needs to read data and doesn't need to write
information, create a database user whose permissions are `read-only`.
Always adjust the database user according to your web applications needs.

## Use a strong password

When creating your database access, choose a strong password. You can use
password managers to generate a strong password.

## Remove default admin passwords

Most DBS have default accounts and most of them have no passwords on their
highest privilege user.

For example, MariaDB, and MongoDB use `root` with no password,

Which means that if there is no password, the attacker could gain access to
everything.

Also, don't forget to remove your credentials and/or private key(s) if you're
going to post your code on a publicly accessible repository in Github.

[1]: https://strongpasswordgenerator.com/
