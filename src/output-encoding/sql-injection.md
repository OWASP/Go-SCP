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

Complete Examples
-----------------

### Basic CRUD Operations

```go
// CREATE - Insert with parameters
func createUser(ctx context.Context, db *sql.DB, user *User) error {
    query := `INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`
    _, err := db.ExecContext(ctx, query, user.Name, user.Email, time.Now())
    return err
}

// READ - Select with parameters
func getUserByID(ctx context.Context, db *sql.DB, id int64) (*User, error) {
    query := `SELECT id, name, email FROM users WHERE id = ?`
    var user User
    err := db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    if err == sql.ErrNoRows {
        return nil, nil // Not found
    }
    return &user, err
}

// UPDATE - Update with parameters
func updateUserEmail(ctx context.Context, db *sql.DB, id int64, email string) error {
    query := `UPDATE users SET email = ?, updated_at = ? WHERE id = ?`
    result, err := db.ExecContext(ctx, query, email, time.Now(), id)
    if err != nil {
        return err
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return errors.New("user not found")
    }
    return nil
}

// DELETE - Delete with parameters
func deleteUser(ctx context.Context, db *sql.DB, id int64) error {
    query := `DELETE FROM users WHERE id = ?`
    _, err := db.ExecContext(ctx, query, id)
    return err
}
```

### IN Clause Queries

IN clauses require special handling:

```go
// VULNERABLE - Don't do this!
func getUsersByIDsInsecure(db *sql.DB, ids []int64) ([]User, error) {
    idsStr := strings.Join(int64sToStrings(ids), ",")
    query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", idsStr) // Dangerous!
    // ...
}

// SECURE - Build parameters dynamically
func getUsersByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]User, error) {
    if len(ids) == 0 {
        return nil, nil
    }

    // Build placeholders: ?, ?, ?
    placeholders := make([]string, len(ids))
    args := make([]interface{}, len(ids))
    for i, id := range ids {
        placeholders[i] = "?"
        args[i] = id
    }

    query := fmt.Sprintf("SELECT id, name, email FROM users WHERE id IN (%s)",
        strings.Join(placeholders, ","))

    rows, err := db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}
```

### Dynamic ORDER BY (Common Mistake)

ORDER BY cannot use placeholders - validate against allowlist:

```go
// VULNERABLE - Never do this!
func getUsersOrderedInsecure(db *sql.DB, orderBy string) ([]User, error) {
    query := fmt.Sprintf("SELECT * FROM users ORDER BY %s", orderBy) // SQL Injection!
    // ...
}

// SECURE - Allowlist validation
var allowedOrderColumns = map[string]string{
    "name":       "name",
    "email":      "email",
    "created_at": "created_at",
    "id":         "id",
}

func getUsersOrdered(ctx context.Context, db *sql.DB, orderBy, direction string) ([]User, error) {
    // Validate order column
    column, ok := allowedOrderColumns[orderBy]
    if !ok {
        column = "id" // Safe default
    }

    // Validate direction
    if direction != "ASC" && direction != "DESC" {
        direction = "ASC" // Safe default
    }

    // Safe to use - validated values only
    query := fmt.Sprintf("SELECT id, name, email FROM users ORDER BY %s %s", column, direction)

    rows, err := db.QueryContext(ctx, query)
    // ...
}
```

### LIKE Queries

Escape wildcard characters in LIKE patterns:

```go
// SECURE - Escape user input for LIKE queries
func searchUsers(ctx context.Context, db *sql.DB, term string) ([]User, error) {
    // Escape SQL LIKE wildcards
    escaped := strings.ReplaceAll(term, "%", "\\%")
    escaped = strings.ReplaceAll(escaped, "_", "\\_")

    query := `SELECT id, name, email FROM users WHERE name LIKE ? ESCAPE '\\'`
    pattern := "%" + escaped + "%"

    rows, err := db.QueryContext(ctx, query, pattern)
    // ...
}
```

### Using Prepared Statements

For repeated queries, use prepared statements:

```go
type UserRepository struct {
    db         *sql.DB
    getByID    *sql.Stmt
    getByEmail *sql.Stmt
}

func NewUserRepository(db *sql.DB) (*UserRepository, error) {
    getByID, err := db.Prepare("SELECT id, name, email FROM users WHERE id = ?")
    if err != nil {
        return nil, err
    }

    getByEmail, err := db.Prepare("SELECT id, name, email FROM users WHERE email = ?")
    if err != nil {
        getByID.Close()
        return nil, err
    }

    return &UserRepository{
        db:         db,
        getByID:    getByID,
        getByEmail: getByEmail,
    }, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.getByID.QueryRowContext(ctx, id).Scan(&user.ID, &user.Name, &user.Email)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) Close() {
    r.getByID.Close()
    r.getByEmail.Close()
}
```

### Using sqlx for Named Parameters

The `sqlx` package provides named parameter support:

```bash
go get github.com/jmoiron/sqlx
```

```go
import "github.com/jmoiron/sqlx"

func createUserNamed(ctx context.Context, db *sqlx.DB, user *User) error {
    query := `INSERT INTO users (name, email, created_at)
              VALUES (:name, :email, :created_at)`

    _, err := db.NamedExecContext(ctx, query, map[string]interface{}{
        "name":       user.Name,
        "email":      user.Email,
        "created_at": time.Now(),
    })
    return err
}
```

ORM Considerations
------------------

ORMs like GORM provide SQL injection protection by default:

```go
import "gorm.io/gorm"

// SECURE - GORM uses parameterized queries internally
func getUserByEmail(db *gorm.DB, email string) (*User, error) {
    var user User
    result := db.Where("email = ?", email).First(&user)
    return &user, result.Error
}

// VULNERABLE - Raw SQL with string concatenation
func getUserInsecure(db *gorm.DB, email string) (*User, error) {
    var user User
    // Don't do this!
    db.Raw("SELECT * FROM users WHERE email = '" + email + "'").Scan(&user)
    return &user, nil
}
```

Security Best Practices Summary
-------------------------------

| Do | Don't |
|----|-------|
| Use parameterized queries | Concatenate user input into queries |
| Validate ORDER BY against allowlist | Use user input directly in ORDER BY |
| Escape LIKE wildcards | Trust ORM raw query methods |
| Use context with timeouts | Ignore query errors |
| Limit result counts | Return unlimited results |

Check the Database Security section in this guide to get more in-depth
information about this topic.

[1]: https://pkg.go.dev/database/sql#DB.Prepare
[2]: https://go.dev/doc/database/sql-injection
[3]: https://bobby-tables.com/go
