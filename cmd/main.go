package main

import (
	"fmt"
	"log"
	"time"

	"github.com/adamhassel/schedule"
	"github.com/robfig/cron"
)

func main() {
	list, err := schedule.Example.PruneNightHours(3)
	if err != nil {
		log.Println(err)
	}
	log.Println("len ", len(schedule.Example))
	c := list.NCheapest(12)
	c.Print()
	fmt.Printf("Total spent on schedule: %.2f", c.Total(500))
}

func parseCron(cronspec string) (time.Time, error) {
	s, err := cron.ParseStandard(cronspec)
	if err != nil {
		return time.Time{}, err
	}
	return s.Next(time.Now()), nil
}
