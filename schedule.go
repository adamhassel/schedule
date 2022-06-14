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

type HourPrice struct {
	Hour  uint
	Price float64
}

// Entry is a full start-stop part of a schedule
type Entry struct {
	Start time.Time
	Stop  time.Time
	Cost  float64
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

// Schedule will compact the hour-list into a shorter list of start and stop times with prices.
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
			se.Cost += hp.Price // Don;t just add the kWh-prices....
			continue
		}
		if se.Stop.Hour() == int(hp.Hour) {
			se.Stop = Hour(today, int(hp.Hour)+1)
			if i != len(h)-1 {
				continue
			}
		}
		schedule = append(schedule, se)
		se = Entry{}
		i--
	}
	return schedule
}

/*// FIXME: Maybe move these guys to the Schellydule package?
type cronjob struct {
	cron    string
	command string
}
*/

//type Cronjobs []cronjob

/*// Fixme: Unused
func (s Schedule) Cron_unused() []cronjob {
	if len(s) == 0 {
		return nil
	}
	rv := make([]cronjob, 0, len(s)*2)
	for _, e := range s {
		start := cronjob{e.Start.Format("04 15 * * *"), "On"}
		stop := cronjob{e.Stop.Format("04 15 * * *"), "Off"}

		rv = append(rv, start, stop)
	}
	return rv
}

*/

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

// Total calculates the total money spent by running the schedule at `consumption` Watts (NOT kW!)
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

func (h HourPrices) NCheapest(n int) HourPrices {
	if n > len(h) {
		n = len(h)
	}
	h.Print()
	sort.Sort(byPrice{h})
	h.Print()
	sub := h[:n]
	fmt.Println(len(sub))
	sort.Sort(byHour{sub})
	return sub
}

// PruneNightHours reduces the number of candidates at night to at most n
func (h HourPrices) PruneNightHours(n int) (HourPrices, error) {
	sunrise, sunset, err := getSunriseSunset(time.Now())
	fmt.Println("sunrise:", sunrise.Format("15:04"), "sunset:", sunset.Format("15:04"))
	if err != nil {
		return nil, err
	}
	sort.Sort(byPrice{h})
	var remove = make([]int, 0, n)
	var night int
	for i, hp := range h {
		if err != nil {
			return nil, err
		}
		if hp.Hour < uint(sunrise.Hour()) || hp.Hour > uint(sunset.Hour()) {
			fmt.Println("removing Hour", hp.Hour, "at index", i)
			night++
			remove = append(remove, i)
		}
	}

	// calculate number of elements to remove
	r := night - n
	if r < 0 {
		r = 0
	}
	if r > len(remove) {
		r = len(remove)
	}
	fmt.Println("Will remove", r)
	remove = remove[0:r]
	sort.Slice(sort.IntSlice(remove), func(i int, j int) bool { return remove[i] > remove[j] })
	for _, i := range remove {
		fmt.Println("removing", i, h[i].Hour)
		h.RemoveAtIdx(i, true)
		fmt.Println("done", i)
	}
	fmt.Println("Len:", len(h))
	return h, nil
}

// RemoveAtIdx removes the element at index i
func (h *HourPrices) RemoveAtIdx(i int, preserveOrder bool) {
	if preserveOrder {
		// Remove the element at index i from a.
		copy((*h)[i:], (*h)[i+1:]) // Shift a[i+1:] left one index.
		(*h)[len(*h)-1] = nil      // Erase last element (write zero value).
		*h = (*h)[:len(*h)-1]      // Truncate slice.
		return
	}
	(*h)[i] = (*h)[len(*h)-1] // Copy last element to index i.
	(*h)[len(*h)-1] = nil     // Erase last element (write zero value).
	*h = (*h)[:len(*h)-1]     // Truncate slice.
}

func getLongLat() (float64, float64, error) {
	r, err := http.Get("https://ipwhois.app/json/")
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
