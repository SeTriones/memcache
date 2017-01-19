package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/SeTriones/memcache"
	//"github.com/pangudashu/memcache"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
)

var (
	port       = flag.String("port", ":8182", "default listen port")
	serverAddr = flag.String("server", "127.0.0.1:11211", "server address")
	user       = flag.String("user", "user", "user name")
	password   = flag.String("password", "password", "password")
	mc         *memcache.Memcache
)

type Reply map[string]interface{}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func Test(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ID := uuid.NewV4().String()
	_, err := mc.Set(ID, ID)
	log.Infof("set id=%s", ID)
	if err != nil {
		log.Errorf("set err=%v", err)
		WriteJSON(w, http.StatusOK, Reply{"id": ID, "msg": "set err", "err": err.Error()})
		return
	}
	val, _, err := mc.Get(ID)
	log.Infof("get id=%s", ID)
	if err != nil {
		log.Errorf("get err=%v", err)
		WriteJSON(w, http.StatusOK, Reply{"id": ID, "msg": "get err", "err": err.Error()})
		return
	}
	v := val.(string)
	if v != ID {
		log.Errorf("val get err,bad val=%v", v)
		WriteJSON(w, http.StatusOK, Reply{"id": ID, "msg": "get err", "err": "value mismatch"})
		return
	}
	WriteJSON(w, http.StatusOK, Reply{"id": ID, "msg": "succ"})
	return
}

func main() {
	flag.Parse()
	router := httprouter.New()
	router.GET("/test", Test)

	var err error
	server := &memcache.Server{Address: *serverAddr, Weight: 100, User: *user, Password: *password, InitConn: 16, MaxConn: 256, IdleTime: time.Second * time.Duration(10)}
	//server := &memcache.Server{Address: *serverAddr, Weight: 100, InitConn: 16, MaxConn: 256, IdleTime: time.Second * time.Duration(10)}
	mc, err = memcache.NewMemcache([]*memcache.Server{server})
	if err != nil {
		panic(err)
	}
	mc.SetTimeout(time.Millisecond*5000, time.Millisecond*2000, time.Millisecond*2000)

	log.Fatal(http.ListenAndServe(*port, router))
}
