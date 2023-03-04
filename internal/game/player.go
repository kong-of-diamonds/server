package game

import (
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/spaolacci/murmur3"
)

type player struct {
	ID    string `msgpack:"id" json:"id"`
	Name  string `msgpack:"name" json:"name"`
	Score int32  `msgpack:"score" json:"score"`
	conn  *websocket.Conn
	readx int
	mu    sync.Mutex
}

func newPlayer(id string, name string, conn *websocket.Conn) *player {
	return &player{
		ID:    id,
		Name:  name,
		Score: 0,
		conn:  conn,
		readx: 0,
	}
}

func (p *player) isDead() bool {
	return p.Score <= -10
}

func (p *player) lost() {
	atomic.AddInt32(&p.Score, -1)
}

func (p *player) sendEvents(es []*event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.readx > len(es)-1 {
		return
	}
	for _, e := range es[p.readx:] {
		if e.outOfDate {
			continue
		}
		data, errMarshallData := e.Serialize()

		if errMarshallData != nil {
			log.Printf("sendEvents Marshall(): %v", errMarshallData)
			return
		}
		if err := p.writeMessage(data); err != nil {
			log.Printf("sendEvents Write(): %v", err)
			return
		}
		if e.Type == eventEndGame {
			p.conn.Close()
			break
		}
		p.readx++
		if p.readx > len(es)-1 {
			break
		}
	}
}

func (p *player) writeMessage(data []byte) error {
	return p.conn.WriteMessage(websocket.TextMessage, data)
}

// MARK: -Game methods

func (g *Game) isSessionFull() bool {
	return len(g.players) == cap(g.players)
}

func (g *Game) AddPlayer(id string, conn *websocket.Conn) (*player, error) {
	g.muPlayers.Lock()
	defer g.muPlayers.Unlock()

	if g.isSessionFull() {
		return nil, errors.New("session is full")
	}

	h := murmur3.New32()
	h.Write([]byte(id))
	name := hex.EncodeToString(h.Sum(nil))[:5]

	var p *player = newPlayer(id, name, conn)
	evGameFound, err := newEventGameFound(g).Serialize()
	if err != nil {
		return nil, err
	}
	if err := p.writeMessage(evGameFound); err != nil {
		return nil, err
	}
	g.players = append(g.players, p)
	g.addEvent(newEventPlayerJoined(p))
	if g.isSessionFull() {
		g.AddTurn()
	}

	println("player added", id)

	return p, nil
}

func (g *Game) RmPlayer(rmp *player) {
	if g.isStarted() {
		return
	}
	g.muPlayers.Lock()
	var newPlayersList []*player = make([]*player, 0, cap(g.players))
	for _, p := range g.players {
		if p.ID != rmp.ID {
			newPlayersList = append(newPlayersList, p)
		}
	}
	g.players = newPlayersList
	g.muPlayers.Unlock()

	g.muEvents.Lock()
	for _, ev := range g.events {
		if !(ev.Type == eventPlayerJoined) || len(ev.Players) == 0 {
			continue
		}
		if ev.Players[0].ID == rmp.ID {
			ev.outOfDate = true
		}
	}
	g.muEvents.Unlock()
	println("player removed", rmp.ID)
}

func (g *Game) getAlivePlayersCount() int {
	var count int
	for _, p := range g.players {
		if !p.isDead() {
			count++
		}
	}
	return count
}
