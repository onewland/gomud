package main

import "mud"

type LifeStage struct {
	StageNo int
	Name string
	StageChangeDelay int
	Pre func(AgingTimeListener)
	Post func(AgingTimeListener)
}

type AgingTimeListener interface {
	mud.TimeListener
	LifeStages() map[int]LifeStage
	Stage() LifeStage
	LastChange() int
	SetStage(LifeStage)
	SetStageChanged(int)
}

func addLs(ls LifeStage, lifeStages map[int]LifeStage) {
	lifeStages[ls.StageNo] = ls
}

func AgeLoop(a AgingTimeListener) {
	for {
		now := <- a.Ping()
		stage := a.Stage()
		if now > (a.LastChange() + stage.StageChangeDelay) {
			if(stage.Post != nil) { 
				a.Stage().Post(a)
			}
	
			if(stage.StageChangeDelay > 0) {
				nextStage := (stage.StageNo + 1)
				a.SetStage(a.LifeStages()[nextStage])
				a.SetStageChanged(now)
			}

			if(a.Stage().Pre != nil) {
				a.Stage().Pre(a)
			}
		}
	}
}