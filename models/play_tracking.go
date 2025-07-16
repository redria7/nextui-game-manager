package models

import "time"

type PlayTrackingAggregate struct {
	Id 				[]int
	Name 			string
	Path			string
	PlayTimeTotal 	int
	PlayCountTotal 	int
	FirstPlayedTime time.Time
	LastPlayedTime 	time.Time
}

type PlayTrackingGranular struct {
	PlayTime	int
	StartTime	int
	UpdateTime	int
}
