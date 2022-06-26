package schedule

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/adamhassel/power"
	"github.com/kelvins/sunrisesunset"
	"github.com/tidwall/gjson"
)

const IPLocationURL = "https://ipwhois.app/json/"

type HourPrice struct {
	Hour  uint
	Price float64
}

// Entry is a full start-stop part of a schedule
type Entry struct {
	Start time.Time `json:"start"`
	Stop  time.Time `json:"stop"`
	Cost  float64   `json:"cost,omitempty"`
}

// Schedule is a complete schedule.
type Schedule []Entry

type HourPrices []*HourPrice

type byPrice struct{ HourPrices }
type byHour struct{ HourPrices }

// example has the next 24 hours' power prices, indexed by Hour

func FPToHourPrices(prices power.FullPrices) HourPrices {
	if len(prices) == 0 {
		return nil
	}
	hp := make(HourPrices, len(prices))
	for i, p := range prices {
		hp[i] = &HourPrice{
			Hour:  uint(p.ValidFrom.Hour()),
			Price: p.TotalIncVAT,
		}
	}
	return hp
}

func (h *HourPrices) Add(hour uint, price float64) {
	*h = append(*h, &HourPrice{hour, price})
}

// Schedule will compact the hour-list into a shorter list of start and stop times with prices per kWh.
func (h HourPrices) Schedule() Schedule {
	var schedule = make(Schedule, 0)
	today := Hour(time.Now(), 0)
	// combine adjacent hours in h
	var se Entry
	for i := 0; i < len(h); i++ {
		hp := h[i]
		if se.Start.IsZero() {
			se.Start = Hour(today, int(hp.Hour))
			se.Stop = Hour(today, int(hp.Hour)+1)
			se.Cost += hp.Price
			continue
		}
		if se.Stop.Hour() == int(hp.Hour) {
			se.Stop = Hour(today, int(hp.Hour)+1)
			se.Cost += hp.Price
			if i != len(h)-1 {
				continue
			}
		}
		schedule = append(schedule, se)
		if i == len(h)-1 {
			break
		}
		se = Entry{}
		i--
	}
	return schedule
}

func (e Entry) String() string {
	return e.Start.Format("15:04") + " - " + e.Stop.Format("15:04")
}

func (s Schedule) String() string {
	var out strings.Builder
	for _, e := range s {
		out.WriteString(e.String() + "\n")
	}
	return out.String()
}

func (s Schedule) Strings() []string {
	out := make([]string, len(s))
	for i, e := range s {
		out[i] = e.String()
	}
	return out
}

func (s Schedule) Map(effect float64) map[string]string {
	if effect == 0 {
		effect = 1000
	}
	out := make(map[string]string, len(s))
	var total float64
	for _, e := range s {
		cost := e.Cost * (effect / 1000)
		out[e.String()] = fmt.Sprintf("%.2f", cost)
		total += cost
	}
	out["total"] = fmt.Sprintf("%.2f", total)
	return out
}

func NewSchedule(cap int) HourPrices {
	hp := make(HourPrices, 0, cap)
	return hp

}

func (h HourPrices) Len() int      { return len(h) }
func (h HourPrices) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (p byPrice) Less(i, j int) bool { return p.HourPrices[i].Price < p.HourPrices[j].Price }
func (p byHour) Less(i, j int) bool  { return p.HourPrices[i].Hour < p.HourPrices[j].Hour }

// byPrice returns a slice of hours sorted by Price
func (h HourPrices) byPrice() {
	sort.Sort(byPrice{h})
}

func (h HourPrices) Print() {
	for i, hp := range h {
		fmt.Printf("idx %d Hours %d - %d: %f\n", i, hp.Hour, hp.Hour+1, hp.Price)
	}
}

// Total calculates the total money spent by running the schedule at `cunsumption` Watts (NOT kW!)
func (h HourPrices) Total(consumption int) float64 {
	var rv float64
	if consumption == 0 {
		consumption = 1000.0
	}
	for _, hp := range h {
		rv += hp.Price * float64(consumption) / 1000.0
	}
	return rv
}

// NCheapest returns the n cheapest hours, with at most nh hours between sunset and sunrise
func (h HourPrices) NCheapest(n int, nh int) (HourPrices, error) {
	if n > len(h) {
		n = len(h)
	}
	h.Print()
	// Sort by price
	sort.Sort(byPrice{h})

	// get the n cheapest, while skipping anything that's at night, once the quota is full
	sub := make(HourPrices, 0, n)
	sunrise, sunset, err := getSunriseSunset(time.Now())
	if err != nil {
		return nil, err
	}
	fmt.Println("sunrise/sunset", sunrise.String(), sunset.String())
	var i int
	for _, hp := range h {
		if len(sub) >= n {
			break
		}
		if dark(hp.Hour, sunrise, sunset) {
			if i < nh-1 {
				sub = append(sub, hp)
				i++
			}
			continue
		}
		sub = append(sub, hp)
	}
	fmt.Println(len(sub))
	sort.Sort(byHour{sub})
	return sub, nil
}

// dark returns true if the hour t is between sunrise and sunset
func dark(t uint, rise, set time.Time) bool {
	return t < uint(rise.Hour()) || t > uint(set.Hour())
}

func getLongLat() (float64, float64, error) {
	r, err := http.Get(IPLocationURL)
	if err != nil {
		return 0, 0, err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, 0, err
	}
	if r.StatusCode == 200 {
		long := gjson.Get(string(body), "longitude").Num
		lat := gjson.Get(string(body), "latitude").Num
		return long, lat, nil
	}
	err = fmt.Errorf("error getting location: %d %s: %s", r.StatusCode, r.Status, string(body))
	return 0, 0, err
}

// Hour will return the time in t redefined to the HH:00:00 in `hour`
func Hour(t time.Time, hour int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
}

// getSunriseSunset returns the sunrise and sunset times for the date in `t`
func getSunriseSunset(t time.Time) (time.Time, time.Time, error) {
	lng, lat, err := getLongLat()
	if err != nil {
		// If we can't geolocate, set sunrise to 6 am and sunset to 6 pm.
		return Hour(t, 6), Hour(t, 18), fmt.Errorf("returning default sunrise/sunsert: %w", err)
	}
	_, offsetSec := t.Zone()
	offsetHr := float64(offsetSec / 60 / 60)
	p := sunrisesunset.Parameters{
		Latitude:  lat,
		Longitude: lng,
		UtcOffset: offsetHr,
		Date:      t,
	}
	return p.GetSunriseSunset()
}
