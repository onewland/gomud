package main

import ("fmt"; "mud"; "mud/simple")

func NewBall(universe *mud.Universe, description string) *simple.PhysicalObject {
	ball := simple.NewPhysicalObject(universe)
	ball.SetDescription(fmt.Sprintf("A %s ball",description))
	ball.SetVisible(true)
	ball.SetCarryable(true)
	ball.SetTextHandles("ball", "red ball")
	return ball
}