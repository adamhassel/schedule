module github.com/adamhassel/schedule

go 1.17

require (
	github.com/adamhassel/power v0.0.0-20220612115315-fa64b5019a64
	github.com/kelvins/sunrisesunset v0.0.0-20210220141756-39fa1bd816d5
	github.com/robfig/cron v1.2.0
	github.com/tidwall/gjson v1.14.0
)

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/rickar/cal/v2 v2.1.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
)

replace github.com/adamhassel/power => ../power
