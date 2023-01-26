# sqan

Dead simple package for scanning SQL rows, compatible with the standard library `database/sql`.

Unlike other libraries, sqan maps structs recursively, this simplifies the scanning of complex queries and allows to handle data in the same table with different (child) objects.

## Install

```
go get -u github.com/GGP1/sqan
```

## Usage

```go
type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Stats Stats `json:"stats"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Stats struct {
	// Other libraries wouldn't map this field
	FriendsCount int `json:"friends_count" db:"friends_count"`
}

// Error handling omitted for simplicity sake
func main() {
	rows, _ := db.Query("SELECT * FROM users")

	var users []User
	_ = sqan.Rows(&users, rows)

	for _, user := range users {
		fmt.Println(user.ID, user.Name, user.Stats.FriendsCount, user.CreatedAt)
	}
	// 1 "Alice" 4 time.Time
	// 2 "Bob"   3 time.Time

	rows, _ := db.Query("SELECT friends_count, created_at FROM users WHERE id=$1", 1)

	var user User
	_ = sqan.Row(&user, rows)
	fmt.Println(user.ID, user.Name, user.Stats.FriendsCount, user.CreatedAt)
	// In some cases, it may be useful to use pointer fields so the empty ones are null
	// 0 "" 4 time.Time
}
```

## Documentation

### Mapping

Unexported fields and struct slices aren't mapped.

The *"db"* tag can be used to map a struct field with an SQL one, if no tag is used, the mapping is done by converting the field's name to lower case.

Objects are mapped only once and the mapping is kept inside a Go map for later use. It is assumed that the number of objects to map is not high enough to cause memory issues.

The `sqan.Row` function takes `sql.Rows` as it's not possible to access the returned columns and map them through `sql.Row`.
