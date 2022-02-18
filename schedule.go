package schedule

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/kelvins/geocoder"
	"github.com/kelvins/sunrisesunset"
	geo "github.com/martinlindhe/google-geolocate"
	"github.com/tidwall/gjson"
)

const apikey = "AIzaSyDXRbQTSdRIUf122Vhp77YXM-8ZvRFt_6c"

type HourPrice struct {
	Hour  uint
	Price float64
}

type HourPrices []*HourPrice

type byPrice struct{ HourPrices }
type byHour struct{ HourPrices }

// example has the next 24 hours' power prices, indexed by hour
var Example = HourPrices{
	{21, 1.82},
	{22, 2.09},
	{23, 1.71},
	{0, 1.75},
	{1, 1.70},
	{2, 1.70},
	{3, 1.70},
	{4, 1.70},
	{5, 1.70},
	{6, 1.78},
	{7, 2.66},
	{8, 2.67},
	{9, 2.67},
	{10, 2.63},
	{11, 2.39},
	{12, 2.20},
	{13, 2.24},
	{14, 2.39},
	{15, 2.32},
	{16, 2.57},
	{17, 3.14},
	{18, 3.13},
	{19, 2.53},
	{20, 1.80},
}

func (h *HourPrices) Add(hour uint, price float64) {
	*h = append(*h, &HourPrice{hour, price})
}

func NewSchedule(cap int) HourPrices {
	hp := make(HourPrices, 0, cap)
	return hp
}

func (h HourPrices) Len() int      { return len(h) }
func (h HourPrices) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (p byPrice) Less(i, j int) bool { return p.HourPrices[i].price < p.HourPrices[j].price }
func (p byHour) Less(i, j int) bool  { return p.HourPrices[i].hour < p.HourPrices[j].hour }

// byPrice returns a slice of hours sorted by price
func (h HourPrices) byPrice() {
	sort.Sort(byPrice{h})
}

func (h HourPrices) Print() {
	for i, hp := range h {
		fmt.Printf("idx %d Hours %d - %d: %f\n", i, hp.hour, hp.hour+1, hp.price)
	}
}

// Total calculates the total money spent by running the schedule at `cunsumption` Watts (NOT kW!)
func (h HourPrices) Total(consumption int) float64 {
	var rv float64
	if consumption == 0 {
		consumption = 1000.0
	}
	for _, hp := range h {
		rv += hp.price * float64(consumption) / 1000.0
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
	lng, lat, err := getLongLat()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, offsetSec := now.Zone()
	offsetHr := float64(offsetSec / 60 / 60)
	p := sunrisesunset.Parameters{
		Latitude:  lat,
		Longitude: lng,
		UtcOffset: offsetHr,
		Date:      now,
	}
	sunrise, sunset, err := p.GetSunriseSunset()
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
		if hp.hour < uint(sunrise.Hour()) || hp.hour > uint(sunset.Hour()) {
			fmt.Println("removing hour", hp.hour, "at index", i)
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
		fmt.Println("removing", i, h[i].hour)
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
	r, err := http.Get("http://ipwhois.app/json/")
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

	// See all Address fields in the documentation
	client := geo.NewGoogleGeo(apikey)
	res, err := client.Geolocate()
	if err == nil {

		return res.Lng, res.Lat, nil
	}
	log.Printf("no result by ip: %s", err)
	geocoder.ApiKey = apikey
	address := geocoder.Address{
		Street:     "Enhøjsvej",
		Number:     25,
		City:       "Allerød",
		Country:    "Denmark",
		PostalCode: "3450",
	}

	// Convert address to location (latitude, longitude)
	location, err := geocoder.Geocoding(address)
	if err != nil {
		return 0, 0, err
	}
	return location.Longitude, location.Latitude, nil
}
