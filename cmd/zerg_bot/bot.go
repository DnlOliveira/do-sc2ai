package main

import (
	"github.com/chippydip/go-sc2ai/api"
	"github.com/chippydip/go-sc2ai/botutil"
	"github.com/chippydip/go-sc2ai/client"
	"github.com/chippydip/go-sc2ai/enums/ability"
	"github.com/chippydip/go-sc2ai/enums/zerg"
	"github.com/chippydip/go-sc2ai/search"
	"log"
)

type bot struct {
	*botutil.Bot

	myStartLocation    api.Point2D
	myNaturalLocation  api.Point2D
	enemyStartLocation api.Point2D

	camera api.Point2D

	builtFistOL bool
	builtNatural bool
	builtFirstGas bool
}

func runAgent(info client.AgentInfo) {
	bot := bot{Bot: botutil.NewBot(info)}
	bot.LogActionErrors()

	bot.init()
	for bot.IsInGame() {
		bot.strategy()
		//bot.tactics()

		if err := bot.Step(1); err != nil {
			log.Print(err)
			break
		}
	}
}

func (bot *bot) init() {
	// My hatchery is on start position
	bot.myStartLocation = bot.Self[zerg.Hatchery].First().Pos2D()
	bot.enemyStartLocation = *bot.GameInfo().GetStartRaw().GetStartLocations()[0]
	bot.camera = bot.myStartLocation

	// Find natural location
	expansions := search.CalculateBaseLocations(bot.Bot, false)
	query := make([]*api.RequestQueryPathing, len(expansions))
	for i, exp := range expansions {
		pos := exp.Location
		query[i] = &api.RequestQueryPathing{
			Start: &api.RequestQueryPathing_StartPos{
				StartPos: &bot.myStartLocation,
			},
			EndPos: &pos,
		}
	}
	resp := bot.Query(api.RequestQuery{Pathing: query})
	best, minDist := -1, float32(256)

	for i, result := range resp.GetPathing() {
		if result.Distance < minDist && result.Distance > 5 {
			best, minDist = i, result.Distance
		}
	}
	bot.myNaturalLocation = expansions[best].Location

	// Send a friendly hello
	bot.Chat("(glhf)")
}

func (bot *bot) strategy() {
	// immediately build first drone at the start of game by checking
	// the drone count and matching it with the initial count
	droneCount := bot.Self.CountAll(zerg.Drone)
	//log.Printf("done_count: %v\n", droneCount)
	if droneCount == 12 {
		if !bot.BuildUnit(zerg.Larva, ability.Train_Drone) {
			return // save up
		}
	}

	// build overlord immediately after first drone
	if droneCount == 13 && !bot.builtFistOL {
		if !bot.BuildUnit(zerg.Larva, ability.Train_Overlord) {
			return // save up
		}
		bot.builtFistOL = true
	}

	// build drones up to 16 count
	if bot.builtFistOL && droneCount < 16 {
		if !bot.BuildUnit(zerg.Larva, ability.Train_Drone) {
			return // save up
		}
	}

	// build hatchery at natural
	if droneCount == 16 && !bot.builtNatural {
		if !bot.BuildUnitAt(zerg.Drone, ability.Build_Hatchery, bot.myNaturalLocation) {
			return
		}
		bot.builtNatural = true
	}

	// build drones to 18, then pool, then to 20, then gas
	if bot.builtNatural && droneCount < 18 {
		if !bot.BuildUnit(zerg.Larva, ability.Train_Drone) {
			return // save up
		}
	} else if bot.builtNatural && droneCount == 18 {
		pos := bot.myStartLocation.Offset(bot.enemyStartLocation, 5)
		if !bot.BuildUnitAt(zerg.Drone, ability.Build_SpawningPool, pos) {
			return // save up
		}
	}

	pool := bot.Self[zerg.SpawningPool].First()
	if !pool.IsNil() && droneCount < 20 && !bot.builtFirstGas {
		if !bot.BuildUnit(zerg.Larva, ability.Train_Drone) {
			return // save up
		}
	} else if !pool.IsNil() && droneCount == 20 && !bot.builtFirstGas {
		if geyser := bot.Neutral.Vespene().CloserThan(10, bot.myStartLocation).Choose(func(u botutil.Unit) bool {
			return bot.Self[zerg.Extractor].CloserThan(1, u.Pos2D()).First().IsNil()
		}).First(); !geyser.IsNil() {
			if !bot.BuildUnitOn(zerg.Drone, ability.Build_Extractor, geyser) {
				return // save up
			}
			bot.builtFirstGas = true
		}
	}

	if bot.builtFirstGas {
		// Build overlords as needed (want at least 3 spare supply per hatch)
		hatches := bot.Self.Count(zerg.Hatchery)
		if bot.FoodLeft() <= 3*hatches && bot.Self.CountInProduction(zerg.Overlord) == 0 {
			if !bot.BuildUnit(zerg.Larva, ability.Train_Overlord) {
				return // save up
			}
		}

		// Get a queen for every hatch if we still have minerals
		bot.BuildUnits(zerg.Hatchery, ability.Train_Queen, hatches-bot.Self.CountAll(zerg.Queen))
	}
}




























