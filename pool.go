package memcache

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

//连接池
type ConnectionPool struct {
	pool     chan *Connection
	address  string
	maxCnt   int
	totalCnt int
	idleTime time.Duration
	user     string
	password string

	sync.Mutex
}

func open(address string, user string, password string, maxCnt int, initCnt int, idelTime time.Duration) (pool *ConnectionPool) {
	pool = &ConnectionPool{
		pool:     make(chan *Connection, maxCnt),
		address:  address,
		maxCnt:   maxCnt,
		idleTime: idelTime,
		user:     user,
		password: password,
	}

	for i := 0; i < initCnt; i++ {
		conn, err := connect(address)
		if err != nil {
			log.Errorf("conn connect fail")
			continue
		}
		err = conn.auth(user, password)
		if err != nil {
			log.Errorf("conn auth fail")
			continue
		}
		pool.totalCnt++
		pool.pool <- conn
	}
	return pool
}

func (this *ConnectionPool) Get() (conn *Connection, err error) {
	for {
		conn, err = this.get()

		if err != nil {
			return nil, err
		}

		if conn.lastActiveTime.Add(this.idleTime).UnixNano() > time.Now().UnixNano() {
			break
		} else {
			this.Release(conn)
		}
	}
	conn.lastActiveTime = time.Now()
	return conn, err
}

func (this *ConnectionPool) get() (conn *Connection, err error) {
	select {
	case conn = <-this.pool:
		return conn, nil
	default:
	}

	this.Lock()

	if this.totalCnt >= this.maxCnt {
		//阻塞，直到有可用连接
		conn = <-this.pool
		this.Unlock()
		return conn, nil
	}

	//create new connect
	conn, err = connect(this.address)
	if err != nil {
		this.Unlock()
		return nil, err
	}
	err = conn.auth(this.user, this.password)
	if err != nil {
		this.Unlock()
		return nil, err
	}
	this.totalCnt++
	this.Unlock()

	return conn, nil
}

func (this *ConnectionPool) Put(conn *Connection) {
	if conn == nil {
		return
	}

	this.pool <- conn
}

func (this *ConnectionPool) Release(conn *Connection) {
	conn.Close()
	this.Lock()
	defer this.Unlock()
	this.totalCnt = this.totalCnt - 1
}

//clear pool
func (this *ConnectionPool) Close() {
	for i := 0; i < len(this.pool); i++ {
		conn := <-this.pool
		conn.Close()
	}
}
