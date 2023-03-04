package game

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

type turnPick struct {
	Player       *player `msgpack:"player" json:"player"`
	PickedNumber float64 `msgpack:"pickedNumber" json:"pickedNumber"`
	Win          bool    `msgpack:"win" json:"win"`
}

type turn struct {
	Picks     []*turnPick `msgpack:"picks" json:"picks"`
	Created   int64       `msgpack:"created" json:"created"`
	Deadline  int64       `msgpack:"deadline" json:"deadline"`
	WinNumber float64     `msgpack:"win_number" json:"win_number"`
	mu        sync.Mutex
}

func newTurn(alivePlayersNumber int, players []*player, timeout time.Duration) *turn {
	var turnPicks []*turnPick = make([]*turnPick, 0, len(players))

	for _, p := range players {
		if p.isDead() {
			continue
		}
		turnPicks = append(turnPicks, &turnPick{Player: p, PickedNumber: -1, Win: false})
	}
	now := time.Now().In(time.UTC)
	return &turn{
		Picks:    turnPicks,
		Created:  now.UnixMicro(),
		Deadline: now.Add(timeout).UnixMicro(),
	}
}

func (t *turn) AddPick(p *player, num float64) error {
	if p.isDead() {
		return errors.New("dead")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, tp := range t.Picks {
		if tp.Player.ID != p.ID {
			continue
		}
		if tp.PickedNumber >= 0 {
			return errors.New("already picked")
		}
		tp.PickedNumber = num
	}
	return nil
}

func (t *turn) computeWinners(factor float64) {
	var sum float64
	var nums []float64 = make([]float64, 0, len(t.Picks))
	for _, tp := range t.Picks {
		if tp.PickedNumber < 0 {
			continue
		}
		sum += tp.PickedNumber
		nums = append(nums, tp.PickedNumber)
	}

	t.WinNumber = sum / float64(len(t.Picks)) * factor

	sort.Float64s(nums)

	if len(nums) == 0 {
		for _, tp := range t.Picks {
			tp.Player.Score = -10
		}
		return
	}

	var closest float64 = nums[0]
	for _, n := range nums {
		if math.Abs(t.WinNumber-n) < math.Abs(t.WinNumber-closest) {
			closest = n
		}
	}

	println("=====================")
	println("closest", closest)
	println("win num", t.WinNumber)
	println("numbers len", len(nums))
	for _, n := range nums {
		print(n, " ")
	}
	println("\n=====================")

	for _, tp := range t.Picks {
		if tp.PickedNumber < 0 {
			tp.Player.Score = -10
			continue
		}
		if tp.PickedNumber == closest {
			tp.Win = true
			continue
		}
		tp.Player.lost()
	}
}

// MARK: -Game methods

func (g *Game) AddTurn() {
	g.muTurns.Lock()
	var t *turn = newTurn(g.getAlivePlayersCount(), g.players, g.timeout)
	g.turns = append(g.turns, t)
	g.muTurns.Unlock()
	g.beginTurn(t)
}

func (g *Game) CurrentTurn() *turn {
	if len(g.turns) == 0 {
		return nil
	}
	return g.turns[len(g.turns)-1]
}

func (g *Game) beginTurn(t *turn) {
	g.addEvent(newEventStartTurn(t))
	time.AfterFunc(g.timeout, g.endTurn)
}

func (g *Game) endTurn() {
	var t *turn = g.CurrentTurn()
	t.computeWinners(g.Factor)
	g.addEvent(newEventEndTurn(t))
	var deadCount int

	for _, tp := range t.Picks {
		if tp.Player.isDead() {
			deadCount++
		}
	}

	if cap(g.players)-deadCount <= 1 {
		println("game end")
		g.endGame()
		return
	}
	g.AddTurn()
}
