package main

import (
	"github.com/chippydip/go-sc2ai/api"
	"github.com/chippydip/go-sc2ai/botutil"
	"github.com/chippydip/go-sc2ai/client"
	"log"
)

type bot struct {
	*botutil.Bot

	myStartLocation    api.Point2D
	myNaturalLocation  api.Point2D
	enemyStartLocation api.Point2D

	camera api.Point2D
}

func runAgent(info client.AgentInfo) {
	bot := bot{Bot: botutil.NewBot(info)}
	bot.LogActionErrors()

	//bot.init()
	for bot.IsInGame() {
		//bot.strategy()
		//bot.tactics()

		if err := bot.Step(1); err != nil {
			log.Print(err)
			break
		}
	}
}
