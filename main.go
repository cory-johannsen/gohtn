package main

import (
	"fmt"
	"github.com/cory-johannsen/gohtn/actor"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"github.com/cory-johannsen/gohtn/loader"
	"log"
	"strings"
	"time"
)

func initializeEngine(cfg *config.Config) *engine.Engine {

	htnEngine := &engine.Engine{
		Actors:        make(actor.Actors),
		Sensors:       make(gohtn.Sensors),
		Actions:       make(engine.Actions),
		Conditions:    make(engine.Conditions),
		TaskResolvers: make(gohtn.TaskResolvers),
		Tasks:         make(gohtn.Tasks),
		Methods:       make(engine.Methods),
		Planner:       nil,
		Domain:        nil,
	}

	vendor := &actor.Vendor{
		NPC: actor.NPC{
			ActorName:     "Vendor",
			ActorLocation: &actor.Point{X: 0.0, Y: 0.0},
		},
		Range: 10.0,
	}
	htnEngine.Actors[vendor.Name()] = vendor
	player := &actor.Player{
		ActorName:     "Player",
		ActorLocation: &actor.Point{X: 20.0, Y: 5.0},
	}
	htnEngine.Actors[player.Name()] = player

	log.Println("loading conditions")
	conditions, err := loader.LoadConditions(cfg)
	if err != nil {
		panic(err)
	}

	conditions["AfterWorkStart"] = &gohtn.ComparisonCondition[int64]{
		Comparison: gohtn.GTE,
		Value:      1,
		Property:   "HourOfDay",
		Comparator: func(value int64, property int64, comparison gohtn.Comparison) bool {
			return property >= value
		},
	}
	conditions["BeforeWorkEnd"] = &gohtn.ComparisonCondition[int64]{
		Comparison: gohtn.LTE,
		Value:      14,
		Property:   "HourOfDay",
		Comparator: func(value int64, property int64, comparison gohtn.Comparison) bool {
			return property <= value
		},
	}
	conditions["CustomerNotEngaged"] = &gohtn.ComparisonCondition[int64]{
		Comparison: gohtn.EQ,
		Value:      0,
		Property:   "CustomersEngaged",
		Comparator: gohtn.Int64Comparator,
	}
	conditions["CustomersInRange"] = &gohtn.ComparisonCondition[int]{
		Comparison: gohtn.GT,
		Value:      0,
		Property:   "CustomersInRange",
		Comparator: gohtn.IntComparator,
	}
	conditions["NoCustomersInRange"] = &gohtn.ComparisonCondition[int]{
		Comparison: gohtn.EQ,
		Value:      0,
		Property:   "CustomersInRange",
		Comparator: gohtn.IntComparator,
	}
	htnEngine.Conditions = conditions

	htnEngine.Conditions["CustomerIsNPC"] = &gohtn.FuncCondition{
		Name: "CustomerIsNPC",
		Evaluator: func(state *gohtn.State) bool {
			// TODO: fetch the current customer for the vendor and check if they are an NPC
			return false
		},
	}
	htnEngine.Conditions["CustomerIsPlayer"] = &gohtn.FuncCondition{
		Name: "CustomerIsNPC",
		Evaluator: func(state *gohtn.State) bool {
			// TODO: fetch the current customer for the vendor and check if they are the player
			return true
		},
	}

	htnEngine.Actions["Wait"] = func(state *gohtn.State) error {
		log.Println("waiting")
		return nil
	}

	htnEngine.Actions["StartWork"] = func(state *gohtn.State) error {
		log.Println("starting work shift")
		return nil
	}
	htnEngine.Actions["EndWork"] = func(state *gohtn.State) error {
		log.Println("ending work shift")
		return nil
	}

	log.Println("loading sensors")

	now := time.Now()
	hourOfDaySensor := &gohtn.HourOfDaySensor{
		TickSensor: gohtn.TickSensor{
			StartedAt:    now,
			TickDuration: 10 * time.Second,
		},
	}
	htnEngine.Sensors["HourOfDay"] = hourOfDaySensor

	customersEngagedSensor := &gohtn.CustomersEngagedSensor{
		Vendor: vendor,
	}
	htnEngine.Sensors["CustomersEngaged"] = customersEngagedSensor

	customersInRangeSensor := &gohtn.CustomersInRangeSensor{
		Vendor: vendor,
		Actors: htnEngine.Actors,
	}
	htnEngine.Sensors["CustomersInRange"] = customersInRangeSensor

	log.Println("loading taskResolvers")
	taskLoader := &loader.TaskLoader{}
	taskResolvers, err := taskLoader.LoadTaskResolvers(cfg, htnEngine)
	if err != nil {
		panic(err)
	}
	log.Printf("Loaded %d taskResolvers", len(taskResolvers))

	log.Println("loading methods")
	methods, err := loader.LoadMethods(cfg, taskLoader, htnEngine)
	if err != nil {
		panic(err)
	}
	htnEngine.Methods = methods

	log.Println("loading task graph")
	taskGraph, err := loader.LoadTaskGraph(cfg, htnEngine)
	if err != nil {
		panic(err)
	}
	htnEngine.Domain = taskGraph

	htnEngine.Planner = &gohtn.Planner{
		Tasks: taskGraph,
	}

	return htnEngine
}

func initializeState(htnEngine *engine.Engine) (*gohtn.State, error) {
	properties := make(map[string]any)
	hourOfDay := &gohtn.Property[int64]{
		Name: "HourOfDay",
		Value: func(state *gohtn.State) int64 {
			sensor, err := state.Sensor("HourOfDay")
			if err != nil {
				log.Fatal(err)
			}
			val, err := sensor.(*gohtn.HourOfDaySensor).Get()
			if err != nil {
				log.Fatal(err)
			}
			return val
		},
	}
	properties["HourOfDay"] = hourOfDay
	customersInRange := &gohtn.Property[int]{
		Name: "CustomersInRange",
		Value: func(state *gohtn.State) int {
			sensor, err := state.Sensor("CustomersInRange")
			if err != nil {
				log.Fatal(err)
			}
			val, err := sensor.(*gohtn.CustomersInRangeSensor).Get()
			if err != nil {
				log.Fatal(err)
			}
			return val
		},
	}
	properties["CustomersInRange"] = customersInRange

	return &gohtn.State{
		Sensors:    htnEngine.Sensors,
		Properties: properties,
	}, nil
}

func main() {
	cfg, err := loader.LoadConfig("config.json")
	if err != nil {
		panic(err)
	}
	log.Println("initializing HTN engine")
	htnEngine := initializeEngine(cfg)

	// Initialize the state from the sensors
	log.Println("initializing state")
	state, err := initializeState(htnEngine)
	if err != nil {
		panic(err)
	}
	var iteration = 0
	var walkingRight = true
	for {
		log.Printf("iteration %d", iteration)
		iteration++
		plan, err := htnEngine.Planner.Plan(state)
		if err != nil {
			panic(err)
		}
		// We are done when the planner can not find and tasks left to execute
		if len(plan) == 0 {
			log.Println("no tasks to execute")
			break
		} else {
			planTasks := make([]string, 0)
			for _, task := range plan {
				planTasks = append(planTasks, fmt.Sprintf("{%s}", task.String()))
			}
			log.Printf("executing plan {%s}", strings.Join(planTasks, ","))
			_, err = gohtn.Execute(plan, state)
			if err != nil {
				panic(err)
			}
		}

		for _, a := range htnEngine.Actors {
			log.Printf("Actor %s: (%f, %f)", a.Name(), a.Location().X, a.Location().Y)
		}

		time.Sleep(time.Second)

		// Have the player walk back and forth from -20,5 to 20,5
		player := htnEngine.Actors["Player"].(*actor.Player)

		if player.Location().X >= 20 {
			walkingRight = false
			player.Location().X = 19
			continue
		}
		if player.Location().X <= -20 {
			walkingRight = true
			player.Location().X = -19
			continue
		}
		if walkingRight {
			player.Location().X += 1
		} else {
			player.Location().X -= 1
		}
	}
}
