package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adamhassel/power"
	"github.com/adamhassel/schedule"
	"github.com/robfig/cron"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlblR5cGUiOiJDdXN0b21lckFQSV9SZWZyZXNoIiwidG9rZW5pZCI6IjgzZjM0NTk3LTIxNmMtNGM4ZC05OTFjLTZiNDQwNDA5MmNlNyIsIndlYkFwcCI6WyJDdXN0b21lckFwaSIsIkN1c3RvbWVyQXBwIl0sImp0aSI6IjgzZjM0NTk3LTIxNmMtNGM4ZC05OTFjLTZiNDQwNDA5MmNlNyIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiUElEOjkyMDgtMjAwMi0yLTY4OTQ5OTM4MTQ0NiIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL2dpdmVubmFtZSI6IkFkYW0gSGFzc2VsYmFsY2ggSGFuc2VuIiwibG9naW5UeXBlIjoiS2V5Q2FyZCIsInBpZCI6IjkyMDgtMjAwMi0yLTY4OTQ5OTM4MTQ0NiIsInR5cCI6IlBPQ0VTIiwidXNlcklkIjoiMTAwNTAwIiwiZXhwIjoxNjc1NjkwMTcxLCJpc3MiOiJFbmVyZ2luZXQiLCJ0b2tlbk5hbWUiOiJkZWZhdWx0IiwiYXVkIjoiRW5lcmdpbmV0In0.YK5H3zpk0Kyq4SYXqHwdUzQnEV2BUZRKrJuBbXhADWU"
const mid = "571313174111451870"

func main() {
	tomorrow := schedule.Hour(time.Now().Add(24*time.Hour), 0)
	prices, err := power.Prices(tomorrow, tomorrow.Add(24*time.Hour), mid, token)
	if err != nil {
		log.Fatal(err)
	}
	list, err := schedule.FPToHourPrices(prices).PruneNightHours(3)
	/*list, err := schedule.Example.PruneNightHours(3)*/
	if err != nil {
		log.Println(err)
	}
	c := list.NCheapest(12)
	c.Print()
	fmt.Printf("Total spent on schedule: %.2f", c.Total(350))
	fmt.Println("Summarized schedule:")
	s := c.Schedule()
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		fmt.Printf("marshaling error: %s", e)
	}
	fmt.Printf(string(b))
}

func parseCron(cronspec string) (time.Time, error) {
	s, err := cron.ParseStandard(cronspec)
	if err != nil {
		return time.Time{}, err
	}
	return s.Next(time.Now()), nil
}
