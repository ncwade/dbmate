package postgres_test

import (
	"database/sql"
	"github.com/adrianmacneil/dbmate/driver/postgres"
	"github.com/stretchr/testify/require"
	"io"
	"net/url"
	"os"
	"testing"
)

func testURL(t *testing.T) *url.URL {
	str := os.Getenv("POSTGRES_PORT")
	require.NotEmpty(t, str, "missing POSTGRES_PORT environment variable")

	u, err := url.Parse(str)
	require.Nil(t, err)

	u.Scheme = "postgres"
	u.User = url.User("postgres")
	u.Path = "/dbmate"
	u.RawQuery = "sslmode=disable"

	return u
}

func mustClose(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(err)
	}
}

func TestCreateDropDatabase(t *testing.T) {
	d := postgres.Driver{}
	u := testURL(t)

	// drop any existing database
	err := d.DropDatabase(u)
	require.Nil(t, err)

	// create database
	err = d.CreateDatabase(u)
	require.Nil(t, err)

	// check that database exists and we can connect to it
	func() {
		db, err := sql.Open("postgres", u.String())
		require.Nil(t, err)
		defer mustClose(db)

		err = db.Ping()
		require.Nil(t, err)
	}()

	// drop the database
	err = d.DropDatabase(u)
	require.Nil(t, err)

	// check that database no longer exists
	func() {
		db, err := sql.Open("postgres", u.String())
		require.Nil(t, err)
		defer mustClose(db)

		err = db.Ping()
		require.NotNil(t, err)
		require.Equal(t, "pq: database \"dbmate\" does not exist", err.Error())
	}()
}

func TestDatabaseExists(t *testing.T) {
	d := postgres.Driver{}
	u := testURL(t)

	// drop any existing database
	err := d.DropDatabase(u)
	require.Nil(t, err)

	// DatabaseExists should return false
	exists, err := d.DatabaseExists(u)
	require.Nil(t, err)
	require.Equal(t, false, exists)

	// create database
	err = d.CreateDatabase(u)
	require.Nil(t, err)

	// DatabaseExists should return true
	exists, err = d.DatabaseExists(u)
	require.Nil(t, err)
	require.Equal(t, true, exists)
}

func TestDatabaseExists_error(t *testing.T) {
	d := postgres.Driver{}
	u := testURL(t)
	u.User = url.User("invalid")

	exists, err := d.DatabaseExists(u)
	require.Equal(t, "pq: role \"invalid\" does not exist", err.Error())
	require.Equal(t, false, exists)
}
