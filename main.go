package main

import (
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"github.com/cory-johannsen/gohtn/loader"
	"log"
	"math/rand"
	"strings"
)

func initializeEngine(cfg *config.Config) *engine.Engine {
	htnEngine := &engine.Engine{
		Actions:       make(engine.Actions),
		Conditions:    make(engine.Conditions),
		TaskResolvers: make(gohtn.TaskResolvers),
		Methods:       make(engine.Methods),
	}

	log.Println("loading conditions")
	conditions, err := loader.LoadConditions(cfg)
	if err != nil {
		panic(err)
	}
	htnEngine.Conditions = conditions

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
	}

	log.Println("loading tasks")
	taskLoader := &loader.TaskLoader{}
	tasks, err := taskLoader.LoadTaskResolvers(cfg, htnEngine)
	if err != nil {
		panic(err)
	}
	log.Printf("Loaded %d tasks", len(tasks))

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

func main() {
	cfg, err := loader.LoadConfig("config.json")
	if err != nil {
		panic(err)
	}
	log.Println("initializing HTN engine")
	htnEngine := initializeEngine(cfg)

	// Initialize the state from the sensors
	log.Println("initializing state")
	state := gohtn.NewState(
		htnEngine.Sensors,
		map[string]gohtn.Property{
			"HourOfDay": func(state *gohtn.State) float64 {
				sensor, err := state.Sensor("HourOfDay")
				if err != nil {
					log.Fatal(err)
				}
				val, err := sensor.Get()
				if err != nil {
					log.Fatal(err)
				}
				return val
			},
			"CustomersInRange": func(state *gohtn.State) float64 {
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
			},
			"CustomersEngaged": func(state *gohtn.State) float64 {
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
			},
		})

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

		// Update the sensors to move the workflow forwards
		iteration++
		iterations, err := state.Sensor("iterations")
		if err != nil {
			panic(err)
		}
		iterations.(*gohtn.SimpleSensor).Set(float64(iteration))

		// flip the flags on different iterations
		alphaFlag, ok := htnEngine.Conditions["alpha"]
		if !ok {
			panic("alpha flag not found")
		}
		if !alphaFlag.(*gohtn.FlagCondition).Value && iteration > 2 {
			alphaFlag.(*gohtn.FlagCondition).Set(true)
		}
		betaFlag, ok := htnEngine.Conditions["beta"]
		if !ok {
			panic("beta flag not found")
		}
		if !betaFlag.(*gohtn.FlagCondition).Value && iteration > 3 {
			betaFlag.(*gohtn.FlagCondition).Set(true)
		}
		gammaFlag, ok := htnEngine.Conditions["gamma"]
		if !ok {
			panic("gamma flag not found")
		}
		if !gammaFlag.(*gohtn.FlagCondition).Value && iteration > 5 {
			gammaFlag.(*gohtn.FlagCondition).Set(true)
		}

		alpha, err := state.Sensor("alpha")
		if err != nil {
			panic(err)
		}
		alphaValue, err := alpha.Get()
		if err != nil {
			panic(err)
		}

		beta, err := state.Sensor("beta")
		if err != nil {
			panic(err)
		}
		betaValue, err := beta.Get()
		if err != nil {
			panic(err)
		}

		gamma, err := state.Sensor("gamma")
		if err != nil {
			panic(err)
		}
		gammaValue, err := gamma.Get()
		if err != nil {
			panic(err)
		}
		alpha.(*gohtn.SimpleSensor).Set(alphaValue + 0.01)
		beta.(*gohtn.SimpleSensor).Set(betaValue + 0.01)
		gamma.(*gohtn.SimpleSensor).Set(gammaValue + 0.01)

		// flip a coin and set the compound task flag
		flip := rand.Intn(2) == 1
		trueFlag, ok := htnEngine.Conditions["compoundTrue"]
		if !ok {
			panic("beta flag not found")
		}
		trueFlag.(*gohtn.FlagCondition).Set(flip)
	}
}
