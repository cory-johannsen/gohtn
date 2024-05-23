package actor

import "math"

type Point struct {
	X float64
	Y float64
}

func Distance(a *Point, b *Point) float64 {
	xTerm := (b.X - a.X) * (b.X - a.X)
	yTerm := (b.Y - a.Y) * (b.Y - a.Y)
	return math.Sqrt(xTerm + yTerm)
}

type Actor interface {
	Name() string
	IsNPC() bool
	Location() *Point
}

type Actors map[string]Actor

type NPC struct {
	ActorName     string
	ActorLocation *Point
}

func (n *NPC) Name() string {
	return n.ActorName
}

func (n *NPC) IsNPC() bool {
	return true
}

func (n *NPC) Location() *Point {
	return n.ActorLocation
}

type Vendor struct {
	NPC
	Customers Actors
	Range     float64
}

func (v *Vendor) AddCustomer(customer Actor) error {
	return nil
}

func (v *Vendor) RemoveCustomer(customer Actor) error {
	return nil
}

func (v *Vendor) IsCustomer(customer Actor) bool {
	return false
}

type Player struct {
	ActorName     string
	ActorLocation *Point
}

func (p *Player) Name() string {
	return p.ActorName
}

func (p *Player) IsNPC() bool {
	return false
}

func (p *Player) Location() *Point {
	return p.ActorLocation
}
