package db

import (
	"database/sql"
	"sync"

	"github.com/mattn/go-sqlite3"
)

const sqliteDriverName = "sqlite3_pornboss"

var registerOnce sync.Once

func registerSQLiteFunctions() string {
	registerOnce.Do(func() {
		sql.Register(sqliteDriverName, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("splitmix64", splitmix64SQL, true)
			},
		})
	})
	return sqliteDriverName
}

func splitmix64SQL(id int64, seed int64) int64 {
	x := uint64(id) + uint64(seed)
	x += 0x9e3779b97f4a7c15
	z := x
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	z = z ^ (z >> 31)
	return int64(z & 0x7fffffffffffffff)
}
