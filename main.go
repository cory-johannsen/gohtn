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
		Actions:    make(engine.Actions),
		Conditions: make(engine.Conditions),
		Tasks:      make(engine.Tasks),
		Methods:    make(engine.Methods),
	}

	conditions, err := loader.LoadConditions(cfg)
	if err != nil {
		panic(err)
	}
	htnEngine.Conditions = conditions

	htnEngine.Actions["alphaAction"] = func(state *gohtn.State) error {
		current, err := state.Property("beta")
		if err != nil {
			return err
		}
		sensor, err := state.Sensor("beta")
		if err != nil {
			return err
		}
		betaSensor := sensor.(*gohtn.SimpleSensor)
		betaSensor.Set(current + 0.10)
		return nil
	}
	htnEngine.Actions["betaAction"] = func(state *gohtn.State) error {
		current, err := state.Property("gamma")
		if err != nil {
			return err
		}
		sensor, err := state.Sensor("gamma")
		if err != nil {
			return err
		}
		gammaSensor := sensor.(*gohtn.SimpleSensor)
		gammaSensor.Set(current + 0.20)
		return nil
	}
	htnEngine.Actions["gammaAction"] = func(state *gohtn.State) error {
		return nil
	}

	sensors, err := loader.LoadSensors(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d sensors", len(sensors))
	for _, sensor := range sensors {
		log.Printf("%s", sensor.String())
	}

	tasks, err := loader.LoadTasks(cfg, htnEngine)
	if err != nil {
		panic(err)
	}
	log.Printf("Loaded %d tasks", len(tasks))

	methods, err := loader.LoadMethods(cfg, htnEngine)
	if err != nil {
		panic(err)
	}
	htnEngine.Methods = methods

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
	htnEngine := initializeEngine(cfg)

	// Initialize the state from the sensors
	state := gohtn.NewState(
		htnEngine.Sensors,
		map[string]gohtn.Property{
			"alpha": func(state *gohtn.State) float64 {
				sensor, err := state.Sensor("alpha")
				if err != nil {
					log.Fatal(err)
				}
				val, err := sensor.Get()
				if err != nil {
					log.Fatal(err)
				}
				return val
			},
			"beta": func(state *gohtn.State) float64 {
				sensor, err := state.Sensor("beta")
				if err != nil {
					log.Fatal(err)
				}
				val, err := sensor.Get()
				if err != nil {
					log.Fatal(err)
				}
				return val
			},
			"gamma": func(state *gohtn.State) float64 {
				sensor, err := state.Sensor("gamma")
				if err != nil {
					log.Fatal(err)
				}
				val, err := sensor.Get()
				if err != nil {
					log.Fatal(err)
				}
				return val
			},
			"iterations": func(state *gohtn.State) float64 {
				sensor, err := state.Sensor("iterations")
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
