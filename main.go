package main

import (
	"fmt"
	"github.com/cory-johannsen/gohtn/gohtn"
	"github.com/cory-johannsen/gohtn/loader"
	"log"
	"math/rand"
	"strings"
)

func main() {
	cfg, err := loader.LoadConfig("config.json")
	sensors, err := loader.LoadSensors(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d sensors", len(sensors))
	for _, sensor := range sensors {
		log.Printf("%s", sensor.String())
	}

	// Initialize the state from the sensors
	state := gohtn.NewState(
		sensors,
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

	alphaFlag := gohtn.FlagCondition{Value: false}
	alphaGTE := gohtn.ComparisonCondition{
		Comparison: gohtn.GTE,
		Value:      0.65,
		Property:   "beta",
	}
	alphaAction := func(state *gohtn.State) error {
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
	alphaTask := gohtn.NewPrimitiveTask("alpha", []gohtn.Condition{
		&alphaFlag,
		&alphaGTE,
	}, alphaAction)

	betaFlag := gohtn.FlagCondition{Value: false}
	betaGTE := gohtn.ComparisonCondition{
		Comparison: gohtn.GTE,
		Value:      0.95,
		Property:   "beta",
	}
	betaAction := func(state *gohtn.State) error {
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
	betaTask := gohtn.NewPrimitiveTask("beta", []gohtn.Condition{
		&betaFlag,
		&betaGTE,
	}, betaAction)

	gammaFlag := gohtn.FlagCondition{Value: false}
	gammaGTE := gohtn.ComparisonCondition{
		Comparison: gohtn.GTE,
		Value:      0.85,
		Property:   "gamma",
	}
	gammaAction := func(state *gohtn.State) error {
		return nil
	}
	gammaTask := gohtn.NewPrimitiveTask("gamma", []gohtn.Condition{
		&gammaFlag,
		&gammaGTE,
	}, gammaAction)

	// Construct a compound task that has 2 methods.  The choice is based on a simple boolean conditions.
	// Place a max iteration counter condition to force task completion
	iterationFlag := &gohtn.ComparisonCondition{
		Comparison: gohtn.LTE,
		Value:      11,
		Property:   "iterations",
	}
	trueFlag := &gohtn.FlagCondition{Value: false}
	falseFlag := &gohtn.NotFlagCondition{
		FlagCondition: *trueFlag,
	}
	trueMethod := gohtn.NewMethod("true", []gohtn.Condition{iterationFlag, trueFlag}, []gohtn.Task{})
	falseMethod := gohtn.NewMethod("false", []gohtn.Condition{iterationFlag, falseFlag}, []gohtn.Task{})
	compoundTask := gohtn.NewCompoundTask("compound", []*gohtn.Method{
		trueMethod,
		falseMethod,
	})

	goal := gohtn.NewGoalTask("goal",
		[]gohtn.TaskCondition{
			{
				Task: alphaTask,
			},
			{
				Task: betaTask,
			},
			{
				Task: gammaTask,
			},
			{
				Task: compoundTask,
			},
		})
	tasks := &gohtn.TaskGraph{
		Root: &gohtn.TaskNode{
			Task: goal,
			Children: []*gohtn.TaskNode{
				{
					Task:     alphaTask,
					Children: []*gohtn.TaskNode{},
				},
				{
					Task: betaTask,
					Children: []*gohtn.TaskNode{
						{
							Task:     gammaTask,
							Children: []*gohtn.TaskNode{},
						},
					},
				},
				{
					Task:     compoundTask,
					Children: []*gohtn.TaskNode{},
				},
			},
		},
	}
	planner := &gohtn.Planner{
		Tasks: tasks,
	}

	var iteration = 0
	for {
		log.Printf("iteration %d", iteration)
		plan, err := planner.Plan(state)
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
		if !alphaFlag.Value && iteration > 2 {
			alphaFlag.Set(true)
		}
		if !betaFlag.Value && iteration > 3 {
			betaFlag.Set(true)
		}
		if !gammaFlag.Value && iteration > 5 {
			gammaFlag.Set(true)
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
		trueFlag.Set(flip)
	}
}
