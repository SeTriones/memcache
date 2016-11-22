package main

import (
	"flag"
	"fmt"
	"github.com/SeTriones/memcache"
	"os"
	"time"
)

var (
	serverAddr = flag.String("server", "127.0.0.1:11211", "server address")
	user       = flag.String("user", "user", "user name")
	password   = flag.String("password", "password", "password")
)

func main() {
	flag.Parse()
	server := &memcache.Server{Address: *serverAddr, Weight: 100, User: *user, Password: *password}
	mc, err := memcache.NewMemcache([]*memcache.Server{server})
	if err != nil {
		panic(err)
	}
	mc.SetTimeout(time.Millisecond*50, time.Millisecond*20, time.Millisecond*20)
	mc.Set("mytest", "pneumonoultramicroscopicsilicovolcanoconiosis")
	val, cas, err := mc.Get("mytest")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "val=%v, cas=%d\n", val, cas)
}
