# gohtn

a golang implementation of an HTN execution engine 

## Background

The goal of this project is to construct the orchestration and execution logic framework to execute a hierarchical task network.

The HTN used is intentional oversimplified in favor of stronger orchestration and execution logic.  As such the network used for testing performs no useful tasks.

## Implementation

The current implementation contains the following simplifications:
1. Sensors are modelled as simple 64 bit floating point values.
2. State properties are a simple passthrough to the sensors.  This allows for easy testing of control flow and evaluation.
3. The HTN implemented for testing purposes is heavily oversimplified and uses primitive flags and value checking as preconditions.
