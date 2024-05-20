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

type Comparison string

const (
	EQ  Comparison = "=="
	NEQ Comparison = "!="
	LT  Comparison = "<"
	LTE Comparison = "<="
	GT  Comparison = ">"
	GTE Comparison = ">="
)

// ComparisonCondition is a condition that is met if the given Property compares to the specified Value using the given Comparison function
type ComparisonCondition struct {
	Comparison Comparison `json:"comparison"`
	Value      float64    `json:"value"`
	Property   string     `json:"property"`
}

func (g *ComparisonCondition) IsMet(state *State) bool {
	value, err := state.Property(g.Property)
	if err != nil {
		return false
	}
	switch g.Comparison {
	case EQ:
		return value >= g.Value
	case NEQ:
		return value != g.Value
	case LT:
		return value < g.Value
	case LTE:
		return value <= g.Value
	case GT:
		return value > g.Value
	case GTE:
		return value >= g.Value
	}
	return false
}

func (g *ComparisonCondition) String() string {
	return fmt.Sprintf("ComparisonCondition: property %s, value %f, comparison %s", g.Property, g.Value, g.Comparison)
}

// PropertyComparisonCondition is a condition that compares to Property values
type PropertyComparisonCondition struct {
	Comparison Comparison `json:"comparison"`
	LHS        string     `json:"lhs"`
	RHS        string     `json:"rhs"`
}

func (p *PropertyComparisonCondition) IsMet(state *State) bool {
	lhs, err := state.Property(p.LHS)
	if err != nil {
		return false
	}
	rhs, err := state.Property(p.RHS)
	if err != nil {
		return false
	}
	switch p.Comparison {
	case EQ:
		return lhs == rhs
	case NEQ:
		return lhs != rhs
	case GT:
		return lhs > rhs
	case GTE:
		return lhs >= rhs
	case LT:
		return lhs < rhs
	case LTE:
		return lhs <= rhs
	}
	return false
}

func (p *PropertyComparisonCondition) String() string {
	return fmt.Sprintf("PropertyComparisonCondition:  %s %s %s", p.LHS, p.Comparison, p.RHS)
}

type LogicalOperator string

const (
	AND LogicalOperator = "AND"
	OR  LogicalOperator = "OR"
	NOT LogicalOperator = "NOT"
	XOR LogicalOperator = "XOR"
)

type LogicalCondition struct {
	Operator    LogicalOperator `json:"operator"`
	LHSProperty string          `json:"lhs"`
	RHSProperty string          `json:"rhs"`
}

func (l *LogicalCondition) IsMet(state *State) bool {
	lhsFloat, err := state.Property(l.LHSProperty)
	if err != nil {
		return false
	}
	lhs := lhsFloat > 0
	rhsFloat, err := state.Property(l.RHSProperty)
	if err != nil {
		if l.Operator == NOT {
			rhsFloat = 0.0
		} else {
			return false
		}
	}
	rhs := rhsFloat > 0

	switch l.Operator {
	case AND:
		return lhs && rhs
	case OR:
		return lhs || rhs
	case NOT:
		return !lhs
	case XOR:
		return (lhs || rhs) && !(lhs && rhs)
	}
	return false
}

func (l *LogicalCondition) String() string {
	return fmt.Sprintf("LogicalCondition: %s %s %s", l.LHSProperty, l.Operator, l.RHSProperty)
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

type Evaluator func(state *State) bool

type FuncCondition struct {
	Name      string    `json:"name"`
	Evaluator Evaluator `json:"evaluator"`
}

func (f *FuncCondition) IsMet(state *State) bool {
	return f.Evaluator(state)
}

func (f *FuncCondition) String() string {
	return fmt.Sprintf("FuncCondition: %s", f.Name)
}
