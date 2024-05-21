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
		Sensors:       make(engine.Sensors),
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
			ActorName: "Vendor",
		},
	}
	htnEngine.Actors[vendor.Name()] = vendor
	player := &actor.Player{
		ActorName: "Player",
	}
	htnEngine.Actors[player.Name()] = player

	log.Println("loading conditions")
	conditions, err := loader.LoadConditions(cfg)
	if err != nil {
		panic(err)
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
	sensors, err := loader.LoadSensors(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d sensors", len(sensors))
	for _, sensor := range sensors {
		log.Printf("%s", sensor.String())
		htnEngine.Sensors[sensor.Name()] = sensor
	}
	now := time.Now()
	htnEngine.Sensors["HourOfDay"] = &gohtn.HourOfDaySensor{
		TickSensor: gohtn.TickSensor{
			StartedAt:    now,
			TickDuration: time.Minute,
		},
	}
	htnEngine.Sensors["CustomersEngaged"] = &gohtn.CustomersEngagedSensor{
		Vendor: vendor,
	}

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
	properties := make(map[string]gohtn.Property)
	properties["HourOfDay"] = func(state *gohtn.State) float64 {
		sensor, err := state.Sensor("HourOfDay")
		if err != nil {
			log.Fatal(err)
		}
		val, err := sensor.Get()
		if err != nil {
			log.Fatal(err)
		}
		return val
	}
	properties["CustomersInRange"] = func(state *gohtn.State) float64 {
		// TODO: fetch and filter the possible customers by range and return the subset
		sensor, err := state.Sensor("CustomersInRange")
		if err != nil {
			log.Fatal(err)
		}
		val, err := sensor.Get()
		if err != nil {
			log.Fatal(err)
		}
		return val
	}
	properties["CustomersEngaged"] = func(state *gohtn.State) float64 {
		// TODO: fetch and filter the possible customers in range and filter to only those that are not engaged
		sensor, err := state.Sensor("CustomersEngaged")
		if err != nil {
			log.Fatal(err)
		}
		val, err := sensor.Get()
		if err != nil {
			log.Fatal(err)
		}
		return val
	}
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
	for {
		log.Printf("iteration %d", iteration)
		plan, err := htnEngine.Planner.Plan(state)
		if err != nil {
			panic(err)
		}
		// We are done when the planner can not find and tasks left to execute
		if len(plan) == 0 {
			log.Println("no tasks to execute")
			break
		}
		planTasks := make([]string, 0)
		for _, task := range plan {
			planTasks = append(planTasks, fmt.Sprintf("{%s}", task.String()))
		}
		log.Printf("executing plan {%s}", strings.Join(planTasks, ","))
		result, err := gohtn.Execute(plan, state)
		if err != nil {
			panic(err)
		}
		log.Printf("state after iteration: %d: %s", iteration, result.String())
		time.Sleep(time.Second)
	}
}
