package schedule

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Contains data used for examples

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

var SampleCheap = HourPrices{
	Example[4],
	Example[6],
	Example[7],
	Example[8],
	Example[9],
	Example[15],
	Example[16],
	Example[18],
	Example[23],
	Example[0],
}

func TestHourPrices_NCheapest(t *testing.T) {
	type args struct {
		n  int
		nh int
	}
	tests := []struct {
		name    string
		h       HourPrices
		args    args
		want    HourPrices
		wantErr bool
	}{
		{
			name: "check len",
			h:    Example,
			args: args{
				n:  10,
				nh: 3,
			},
			want: HourPrices{
				Example[4],
				Example[6],
				Example[7],
				Example[8],
				Example[9],
				Example[15],
				Example[16],
				Example[18],
				Example[23],
				Example[0],
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.NCheapest(tt.args.n, tt.args.nh)
			if (err != nil) != tt.wantErr {
				t.Errorf("NCheapest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Len(t, got, tt.args.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NCheapest() got = %v, want %+v", got, tt.want)
				got.Print()
			}
		})
	}
}

func SumPrice(t *testing.T, h HourPrices, start, n int) float64 {
	if len(h) < start-1 {
		t.Fatal("h not long enough!")
	}
	var res float64
	for i := start; i < start+n; i++ {
		res += h[i].Price
	}
	return res
}

func TestHourPrices_Schedule(t *testing.T) {
	today := time.Now()
	tests := []struct {
		name string
		h    HourPrices
		want Schedule
	}{
		{
			name: "check that it werks",
			h: HourPrices{
				Example[4],
				Example[6],
				Example[7],
				Example[8],
				Example[9],
				Example[15],
				Example[16],
				Example[18],
				Example[23],
				Example[0],
			},
			want: Schedule{
				{
					Start: Hour(today, int(Example[4].Hour)),
					Stop:  Hour(today, int(Example[4].Hour+1)),
					Cost:  SumPrice(t, Example, 4, 1),
				},
				{
					Start: Hour(today, int(Example[6].Hour)),
					Stop:  Hour(today, int(Example[6].Hour+4)),
					Cost:  SumPrice(t, Example, 6, 4),
				},
				{
					Start: Hour(today, int(Example[15].Hour)),
					Stop:  Hour(today, int(Example[15].Hour+2)),
					Cost:  SumPrice(t, Example, 15, 2),
				},
				{
					Start: Hour(today, int(Example[18].Hour)),
					Stop:  Hour(today, int(Example[18].Hour+1)),
					Cost:  SumPrice(t, Example, 18, 1),
				},
				{
					Start: Hour(today, int(Example[23].Hour)),
					Stop:  Hour(today, int(Example[23].Hour+2)),
					Cost:  Example[23].Price + Example[0].Price,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.h.Schedule()
			assert.Len(t, res, 5)
			assert.Equal(t, len(tt.h), res.Hours())
			assert.Equalf(t, tt.want, res, "Schedule()")
		})
	}
}

func TestHourPrices_DurationHours(t *testing.T) {
	type args struct {
		start time.Time
	}
	tests := []struct {
		name string
		h    HourPrices
		args args
		want int
	}{
		{
			name: "check whatev",
			h:    SampleCheap,
			args: args{
				start: Hour(time.Now(), 3),
			},
			want: 4,
		},
		{
			name: "check end",
			h:    SampleCheap,
			args: args{
				start: Hour(time.Now(), 20),
			},
			want: 2,
		},
		{
			name: "check single",
			h:    SampleCheap,
			args: args{
				start: Hour(time.Now(), 1),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.h.DurationHours(tt.args.start), "DurationHours(%v)", tt.args.start)
		})
	}
}
