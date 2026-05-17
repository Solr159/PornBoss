package db

import (
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/glebarez/go-sqlite"
)

var registerOnce sync.Once

func registerSQLiteFunctions() {
	registerOnce.Do(func() {
		sqlite.MustRegisterDeterministicScalarFunction("splitmix64", 2, scalarInt64Binary(splitmix64SQL))
		sqlite.MustRegisterDeterministicScalarFunction("stable_random_rank", 2, scalarInt64Binary(stableRandomRankSQL))
	})
}

func scalarInt64Binary(fn func(id, seed int64) int64) func(*sqlite.FunctionContext, []driver.Value) (driver.Value, error) {
	return func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		id, err := scalarArgInt64(args, 0)
		if err != nil {
			return nil, err
		}
		seed, err := scalarArgInt64(args, 1)
		if err != nil {
			return nil, err
		}
		return fn(id, seed), nil
	}
}

func scalarArgInt64(args []driver.Value, idx int) (int64, error) {
	if idx >= len(args) {
		return 0, fmt.Errorf("sqlite scalar: missing argument %d", idx)
	}
	switch v := args[idx].(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("sqlite scalar: argument %d: want integer, got %T", idx, args[idx])
	}
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

func stableRandomRankSQL(id int64, seed int64) int64 {
	x := uint64(seed) ^ 0x9e3779b97f4a7c15
	y := uint64(id) + 0xbf58476d1ce4e5b9
	x ^= y + 0x9e3779b97f4a7c15 + (x << 6) + (x >> 2)
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return int64(x & 0x7fffffffffffffff)
}
