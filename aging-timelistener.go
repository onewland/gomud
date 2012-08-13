package main

import "mud"

type LifeStage struct {
	StageNo int
	Name string
	StageChangeDelay int
	Pre func(interface{})
	Post func(interface{})
}

type AgingTimeListener interface {
	mud.TimeListener
	LifeStages() []LifeStage
}

func addLs(ls LifeStage, lifeStages map[int]LifeStage) {
	lifeStages[ls.StageNo] = ls
}