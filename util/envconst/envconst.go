package envconst

import (
	"os"
	"strconv"
	"sync"
	"time"
)

var cache sync.Map

// capture the actual envconst values used at runtime
func DebugDump() map[string]interface{} {
	m := make(map[string]interface{})
	cache.Range(func(k, v interface{}) bool {
		m[k.(string)] = v
		return true
	})
	return m
}

func Duration(varname string, def time.Duration) time.Duration {
	if v, ok := cache.Load(varname); ok {
		return v.(time.Duration)
	}
	e := os.Getenv(varname)
	if e == "" {
		return def
	}
	d, err := time.ParseDuration(e)
	if err != nil {
		panic(err)
	}
	cache.Store(varname, d)
	return d
}

func Int(varname string, def int) int {
	if v, ok := cache.Load(varname); ok {
		return v.(int)
	}
	e := os.Getenv(varname)
	if e == "" {
		return def
	}
	d64, err := strconv.ParseInt(e, 10, strconv.IntSize)
	if err != nil {
		panic(err)
	}
	d := int(d64)
	cache.Store(varname, d)
	return d
}

func Int64(varname string, def int64) int64 {
	if v, ok := cache.Load(varname); ok {
		return v.(int64)
	}
	e := os.Getenv(varname)
	if e == "" {
		return def
	}
	d, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		panic(err)
	}
	cache.Store(varname, d)
	return d
}

func Bool(varname string, def bool) bool {
	if v, ok := cache.Load(varname); ok {
		return v.(bool)
	}
	e := os.Getenv(varname)
	if e == "" {
		return def
	}
	d, err := strconv.ParseBool(e)
	if err != nil {
		panic(err)
	}
	cache.Store(varname, d)
	return d
}
