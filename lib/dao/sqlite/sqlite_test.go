package sqlite

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/mattn/go-sqlite3"

	"github.com/lack-io/vine/lib/dao"
)

func TestDialector(t *testing.T) {
	// This is the DSN of the in-memory SQLite database for these tests.
	const InMemoryDSN = "file:testdatabase?mode=memory&cache=shared"
	// This is the custom SQLite driver name.
	const CustomDriverName = "my_custom_driver"

	// Register the custom SQlite3 driver.
	// It will have one custom function called "my_custom_function".
	sql.Register(CustomDriverName,
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				// Define the `concat` function, since we use this elsewhere.
				err := conn.RegisterFunc(
					"my_custom_function",
					func(arguments ...interface{}) (string, error) {
						return "my-result", nil // Return a string value.
					},
					true,
				)
				return err
			},
		},
	)

	rows := []struct {
		description  string
		dialect      dao.Dialect
		openSuccess  bool
		query        string
		querySuccess bool
	}{
		{
			description:  "Default driver",
			dialect:      NewDialect(dao.DSN(InMemoryDSN)),
			openSuccess:  true,
			query:        "SELECT 1",
			querySuccess: true,
		},
		{
			description:  "Explicit default driver",
			dialect:      NewDialect(dao.DSN(InMemoryDSN)),
			openSuccess:  true,
			query:        "SELECT 1",
			querySuccess: true,
		},
		{
			description: "Bad driver",
			dialect:     NewDialect(dao.DSN(InMemoryDSN), DriverName("not-a-real-driver")),
			openSuccess: false,
		},
		{
			description:  "Custom driver",
			dialect:      NewDialect(dao.DSN(InMemoryDSN), DriverName(CustomDriverName)),
			openSuccess:  true,
			query:        "SELECT 1",
			querySuccess: true,
		},
		{
			description:  "Custom driver, custom function",
			dialect:      NewDialect(dao.DSN(InMemoryDSN), DriverName(CustomDriverName)),
			openSuccess:  true,
			query:        "SELECT my_custom_function()",
			querySuccess: true,
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d/%s", rowIndex, row.description), func(t *testing.T) {
			err := row.dialect.Init()
			if !row.openSuccess {
				if err == nil {
					t.Errorf("Expected Open to fail.")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected Open to succeed; got error: %v", err)
			}
			//if db == nil {
			//	t.Errorf("Expected db to be non-nil.")
			//}
			if row.query != "" {
				err = row.dialect.NewTx().Exec(row.query).Error
				if !row.querySuccess {
					if err == nil {
						t.Errorf("Expected query to fail.")
					}
					return
				}

				if err != nil {
					t.Errorf("Expected query to succeed; got error: %v", err)
				}
			}
		})
	}
}
