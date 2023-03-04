package game

import (
	"encoding/json"

	"github.com/google/uuid"
)

type eventType uint8

const (
	eventGameFound eventType = iota

	eventPlayerJoined eventType = iota
	eventPlayerLeft   eventType = iota

	eventStartTurn eventType = iota
	eventEndTurn   eventType = iota

	eventEndGame eventType = iota
)

type event struct {
	ID        string    `msgpack:"id" json:"id"`
	Type      eventType `msgpack:"type" json:"type"`
	Players   []*player `msgpack:"players,omitempty" json:"players,omitempty"`
	Turn      *turn     `msgpack:"turn,omitempty" json:"turn,omitempty"`
	Game      *Game     `msgpack:"game,omitempty" json:"game,omitempty"`
	outOfDate bool
}

func newEvent(typ eventType, p []*player, t *turn, g *Game) *event {
	return &event{
		ID:      uuid.NewString(),
		Players: p,
		Turn:    t,
		Type:    typ,
		Game:    g,
	}
}

func (e *event) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

func newEventGameFound(g *Game) *event {
	return newEvent(eventGameFound, nil, nil, g)
}

func newEventPlayerJoined(p *player) *event {
	return newEvent(eventPlayerJoined, []*player{p}, nil, nil)
}

func newEventPlayerLeft(p *player) *event {
	return newEvent(eventPlayerJoined, []*player{p}, nil, nil)
}

func newEventEndTurn(t *turn) *event {
	return newEvent(eventEndTurn, nil, t, nil)
}

func newEventStartTurn(t *turn) *event {
	return newEvent(eventStartTurn, nil, t, nil)
}

func newEventEndGame(ps []*player) *event {
	return newEvent(eventEndGame, ps, nil, nil)
}

// MARK: -Game methods

func (g *Game) addEvent(e *event) {
	g.muEvents.Lock()
	defer g.muEvents.Unlock()
	g.events = append(g.events, e)
	for _, p := range g.players {
		go p.sendEvents(g.events)
	}
}
