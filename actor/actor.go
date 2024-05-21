package actor

type Actor interface {
	Name() string
	IsNPC() bool
}

type Actors map[string]Actor

type NPC struct {
	ActorName string
}

func (n *NPC) Name() string {
	return n.ActorName
}

func (n *NPC) IsNPC() bool {
	return true
}

type Vendor struct {
	NPC
}

type Player struct {
	ActorName string
}

func (p *Player) Name() string {
	return p.ActorName
}

func (p *Player) IsNPC() bool {
	return false
}
