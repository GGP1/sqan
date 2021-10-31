# sqan

Dead simple package for scanning SQL rows, compatible with the standard library `database/sql`.

Unlike other libraries, sqan maps structs recursively, this simplifies the scanning of complex queries and allows to handle data in the same table with different (child) objects.

## Install

```
go get github.com/GGP1/sqan
```

## Usage

```go
package main

import (
	"database/sql"
	"log"

	"github.com/GGP1/sqan"

	_ "github.com/lib/pq"
)

type User struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Stats Stats `json:"stats"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Stats struct {
	FriendsCount int `json:"friends_count" db:"friends_count"` // Other libraries wouldn't map this field
}

func main() {
	db, err := sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id serial PRIMARY KEY,
		name varchar(64),
		friends_count integer,
		created_at timestamp with time zone DEFAULT NOW()
	)`)
	db.Exec("INSERT INTO users (name, friends_count) VALUES ($1, $2), ($3, $4)", "Alice", 4, "Bob", 3)

	users, err := getUsers(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		fmt.Println(user.ID, user.Name, user.Stats.FriendsCount, user.CreatedAt)
	}
	// 1 Alice 4 time
	// 2 Bob 3 time

	user, err := getUser(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user.ID, user.Name, user.Stats.FriendsCount, user.CreatedAt)
	// 0 "" 4 time
}

func getUsers(db *sql.DB) ([]Users, error) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	var users []User
	if err := sqan.Rows(&users, rows); err != nil {
		return nil, err
	}

	return users, nil
}

func getUser(db *sql.DB) (User, error) {
	rows, err := db.Query("SELECT friends_count, created_at FROM users WHERE id=$1", "1")
	if err != nil {
		return User{}, err
	}

	var user User
	if err := sqan.Row(&user, rows); err != nil {
		return User{}, err
	}

	return user, nil
}
```

## Documentation

### Mapping

Unexported fields and struct slices aren't mapped.

The *"db"* tag can be used to map a struct field with an SQL one, if no tag is used, the mapping is done by converting the field's name to lower case.

Objects are mapped only once and the mapping is kept inside a Go map for later use. It is assumed that the number of objects to map is not high enough to cause memory issues.

The `sqan.Row` function takes `sql.Rows` as it's not possible to access the returned columns and map them through `sql.Row`.
