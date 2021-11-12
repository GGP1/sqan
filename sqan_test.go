package sqan

import (
	"database/sql"
	"log"
	"os"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

// Test struct.
type Test struct {
	Letter    string
	Lowercase bool `db:"lower_case"`
	Weight    int
	Skip      []Sub
	Sub       Sub
}

// Sub struct.
type Sub struct {
	Exported   bool
	unexported string
}

var db *sql.DB

var records = []Test{
	{Letter: "A", Lowercase: false, Weight: 100, Sub: Sub{Exported: true}},
	{Letter: "b", Lowercase: true, Weight: 0, Sub: Sub{Exported: false}},
	{Letter: "C", Lowercase: false, Weight: 200, Sub: Sub{Exported: true}},
}

func TestMain(m *testing.M) {
	// docker run -d -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust postgres:14.0-alpine
	sql, err := sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db = sql

	_, _ = db.Exec("DROP TABLE tests")

	createTable := `
CREATE TABLE tests
(
	letter text,
	lower_case bool,
	weight integer,
	exported bool
)`
	if _, err := db.Exec(createTable); err != nil {
		log.Fatal(err)
	}

	q := "INSERT INTO tests (letter, lower_case, weight, exported) VALUES ($1, $2, $3, $4)"
	for _, r := range records {
		if _, err := db.Exec(q, r.Letter, r.Lowercase, r.Weight, r.Sub.Exported); err != nil {
			log.Fatal(err)
		}
	}

	os.Exit(m.Run())

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestRows(t *testing.T) {
	t.Run("Struct slice", func(t *testing.T) {
		rows, err := db.Query("SELECT letter, weight, lower_case, exported FROM tests")
		if err != nil {
			t.Fatal(err)
		}

		var got []Test
		if err := Rows(&got, rows); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(records, got) {
			t.Errorf("Expected %v, got %v", records, got)
		}
	})

	t.Run("String slice", func(t *testing.T) {
		expected := []string{"A", "b", "C"}
		rows, err := db.Query("SELECT letter FROM tests")
		if err != nil {
			t.Fatal(err)
		}

		var got []string
		if err := Rows(&got, rows); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, got) {
			t.Errorf("Expected %v, got %v", records, got)
		}
	})
}

func TestRowsErrors(t *testing.T) {
	rows, err := db.Query("SELECT 1 FROM tests")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc string
		dest interface{}
	}{
		{
			desc: "Not a pointer",
			dest: []Test{},
		},
		{
			desc: "Not nil",
			dest: nil,
		},
		{
			desc: "Not a slice",
			dest: &Test{},
		},
		{
			desc: "Not a struct slice",
			dest: &[]int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := Rows(tc.dest, rows); err == nil {
				t.Fatal("Expected an error and got nil")
			}
		})
	}
}

func TestRow(t *testing.T) {
	target := records[0]
	expected := Test{Weight: target.Weight, Sub: target.Sub}
	rows, err := db.Query("SELECT weight, exported FROM tests WHERE letter=$1", target.Letter)
	if err != nil {
		t.Fatal(err)
	}

	var got Test
	if err := Row(&got, rows); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestRowErrors(t *testing.T) {
	rows, err := db.Query("SELECT 1 FROM tests")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc string
		dest interface{}
	}{
		{
			desc: "Not a pointer",
			dest: Test{},
		},
		{
			desc: "Not nil",
			dest: nil,
		},
		{
			desc: "Not a struct",
			dest: "text",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := Row(tc.dest, rows); err == nil {
				t.Fatal("Expected an error and got nil")
			}
		})
	}
}
