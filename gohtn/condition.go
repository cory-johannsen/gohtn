package gohtn

import "fmt"

type Condition interface {
	IsMet(state *State) bool
	String() string
}

// FlagCondition is a simple condition that is gated by a boolean Value that can be set
type FlagCondition struct {
	Value bool `json:"value"`
}

func (f *FlagCondition) IsMet(state *State) bool {
	return f.Value
}

func (f *FlagCondition) Set(value bool) {
	f.Value = value
}

func (f *FlagCondition) String() string {
	return fmt.Sprintf("FlagCondition: %t", f.Value)
}

// NotFlagCondition embeds a FlagCondition and inverts the behavior
type NotFlagCondition struct {
	FlagCondition `json:"flag_condition"`
}

func (n *NotFlagCondition) IsMet(state *State) bool {
	return !n.FlagCondition.IsMet(state)
}

func (n *NotFlagCondition) String() string {
	return fmt.Sprintf("NotFlagCondition: %t", n.FlagCondition.Value)
}

// GTECondition is a condition that is met if the given Property is GTE the specified Value
type GTECondition struct {
	Value    float64 `json:"value"`
	Property string  `json:"property"`
}

func (g *GTECondition) IsMet(state *State) bool {
	value, err := state.Property(g.Property)
	if err != nil {
		return false
	}
	return value >= g.Value
}

func (g *GTECondition) String() string {
	return fmt.Sprintf("GTECondition: property %s, value %f", g.Property, g.Value)
}

type LTECondition struct {
	Value    float64 `json:"value"`
	Property string  `json:"property"`
}

func (l *LTECondition) IsMet(state *State) bool {
	value, err := state.Property(l.Property)
	if err != nil {
		return false
	}
	return value <= l.Value
}

func (l *LTECondition) String() string {
	return fmt.Sprintf("LTECondition: property %s, value %f", l.Property, l.Value)
}

// TaskCondition is a condition that is met when the given Task is complete
type TaskCondition struct {
	Task Task `json:"task"`
}

func (t *TaskCondition) IsMet(state *State) bool {
	return t.Task.IsComplete()
}

func (t *TaskCondition) String() string {
	return fmt.Sprintf("TaskCondition: %s, complete: %t", t.Task.Name(), t.Task.IsComplete())
}
