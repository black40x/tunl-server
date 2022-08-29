package server

import (
	"errors"
	"github.com/black40x/tunl-core/commands"
	"github.com/black40x/tunl-core/tunl"
	"github.com/black40x/tunl-server/cmd/utils"
	"net"
	"strings"
	"sync"
)

const responseKeyPrefix = "response:"
const bodyChunkKeyPrefix = "body-chunk:"

var ErrorIdBusy = errors.New("connection id is busy")

type ConnectionPool struct {
	pool       sync.Map
	channels   sync.Map
	allowed    sync.Map
	prefixSize int
}

func NewConnectionPool(prefixSize int) *ConnectionPool {
	return &ConnectionPool{
		pool:       sync.Map{},
		channels:   sync.Map{},
		allowed:    sync.Map{},
		prefixSize: prefixSize,
	}
}

func (p *ConnectionPool) generateID(conn *tunl.TunlConn) string {
	remoteIP := conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String()
	remoteIP = strings.ReplaceAll(remoteIP, ".", "-")

	return remoteIP + "-" + utils.RandomString(p.prefixSize)
}

func (p *ConnectionPool) Push(conn *tunl.TunlConn) error {
	id := p.generateID(conn)
	if _, ok := p.pool.Load(id); ok {
		return ErrorIdBusy
	}

	conn.ID = id
	p.pool.Store(id, conn)

	return nil
}

func (p *ConnectionPool) Count() int {
	count := 0
	p.pool.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (p *ConnectionPool) Drop(id string) {
	conn, ok := p.pool.Load(id)
	if ok {
		conn.(*tunl.TunlConn).Close()
		p.pool.Delete(id)
	}

	_, ok = p.allowed.Load(id)
	if ok {
		p.allowed.Delete(id)
	}
}

func (p *ConnectionPool) Get(id string) *tunl.TunlConn {
	conn, ok := p.pool.Load(id)
	if !ok {
		return nil
	}
	return conn.(*tunl.TunlConn)
}

func (p *ConnectionPool) GetResponseChan(uuid string) chan *commands.HttpResponse {
	response, ok := p.channels.Load(responseKeyPrefix + uuid)
	if !ok {
		return nil
	}

	return response.(chan *commands.HttpResponse)
}

func (p *ConnectionPool) GetBodyChunkChan(uuid string) chan *commands.BodyChunk {
	response, ok := p.channels.Load(bodyChunkKeyPrefix + uuid)
	if !ok {
		return nil
	}

	return response.(chan *commands.BodyChunk)
}

func (p *ConnectionPool) MakeResponseChan(uuid string) chan *commands.HttpResponse {
	channel := make(chan *commands.HttpResponse, 100)
	p.channels.Store(responseKeyPrefix+uuid, channel)
	return channel
}

func (p *ConnectionPool) MakeBodyChunkChan(uuid string) chan *commands.BodyChunk {
	channel := make(chan *commands.BodyChunk, 100)
	p.channels.Store(bodyChunkKeyPrefix+uuid, channel)
	return channel
}

func (p *ConnectionPool) SetAllowed(id string) {
	p.allowed.Store(id, true)
}

func (p *ConnectionPool) IsAllowed(id string) bool {
	allow, ok := p.allowed.Load(id)
	if !ok {
		return false
	}

	return allow.(bool)
}

func (p *ConnectionPool) CloseChannels(uuid string) {
	if response := p.GetResponseChan(uuid); response != nil {
		close(response)
		p.channels.Delete(responseKeyPrefix + uuid)
	}
	if bodyChunk := p.GetBodyChunkChan(uuid); bodyChunk != nil {
		close(bodyChunk)
		p.channels.Delete(bodyChunkKeyPrefix + uuid)
	}
}
