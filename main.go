package main

import (
	"fmt"
	"github.com/cory-johannsen/gohtn/gohtn"
	"log"
)

func main() {
	// Define three simple sensors that will be used to feed the state
	alpha := gohtn.NewSimpleSensor(0.5)
	beta := gohtn.NewSimpleSensor(0.5)
	gamma := gohtn.NewSimpleSensor(0.5)
	// Initialize the state from the sensors
	state := gohtn.NewState(
		[]gohtn.Sensor{
			alpha,
			beta,
			gamma,
		},
		map[string]gohtn.Sensor{
			"alpha": alpha,
			"beta":  beta,
			"gamma": gamma,
		})

	alphaFlag := gohtn.FlagCondition{Value: false}
	alphaGTE := gohtn.GTECondition{
		Value:    0.65,
		Property: "beta",
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
	betaGTE := gohtn.GTECondition{
		Value:    0.95,
		Property: "beta",
	}
	betaAction := func(state *gohtn.State) error {
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
	betaTask := gohtn.NewPrimitiveTask("beta", []gohtn.Condition{
		&betaFlag,
		&betaGTE,
	}, betaAction)

	gammaFlag := gohtn.FlagCondition{Value: false}
	gammaGTE := gohtn.GTECondition{
		Value:    0.85,
		Property: "gamma",
	}
	gammaAction := func(state *gohtn.State) error {
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
	gammaTask := gohtn.NewPrimitiveTask("gamma", []gohtn.Condition{
		&gammaFlag,
		&gammaGTE,
	}, gammaAction)

	goal := gohtn.NewGoalTask(
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
		result, err := gohtn.Execute(plan, state)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)

		if goal.IsComplete() {
			log.Println("goal reached")
			break
		}

		iteration++

		// flip the flags on different iterations
		if !alphaFlag.Value && iteration > 3 {
			alphaFlag.Set(true)
		}
		if !betaFlag.Value && iteration > 9 {
			betaFlag.Set(true)
		}
		if !alphaFlag.Value && iteration > 6 {
			gammaFlag.Set(true)
		}

		// increment the sensors to allow the plan to progress towards the goal
		alphaValue, err := alpha.Value()
		if err != nil {
			panic(err)
		}
		betaValue, err := beta.Value()
		if err != nil {
			panic(err)
		}
		gammaValue, err := gamma.Value()
		if err != nil {
			panic(err)
		}
		alpha.Set(alphaValue + 0.01)
		beta.Set(betaValue + 0.01)
		gamma.Set(gammaValue + 0.01)
	}
}
