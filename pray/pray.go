package pray

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hablullah/go-prayer"
	"github.com/asim/mu"
)

var (
	template = `
  <style>
  #name {
    text-align: right;
    width: 100px;
    margin-right: 20px;
    display: inline-block;
  }
  #shahada {
    padding: 50px 20px;
    text-align: center;
    max-width: 600px;
  }
  #times {
    max-width: 600px;
    text-align: center;
    font-size: 2em;
    font-style: italic;
    padding-top: 50px;
  }
  </style>
  <div id="times">
  %s
  </div>
`
)

type City struct {
	Name     string
	Lat, Lon float64
	Location string
}

var cities = []City{
	City{Name: "London", Lat: 51.41334, Lon: -0.36701, Location: "Europe/London"},
	City{Name: "Edinburgh", Lat: 55.9425285, Lon: -3.295193, Location: "Europe/London"},
	City{Name: "Mecca", Lat: 21.422487, Lon: 39.826206, Location: "Asia/Riyadh"},
	City{Name: "Doha", Lat: 25.286106, Lon: 51.534817, Location: "Asia/Qatar"},
	City{Name: "New York", Lat: 40.730610, Lon: -73.935242, Location: "America/New_York"},
	City{Name: "Sydney", Lat: -33.865143, Lon: 151.209900, Location: "Australia/Sydney"},
	City{Name: "Tokyo", Lat: 35.672855, Lon: 139.817413, Location: "Asia/Tokyo"},
}

func dateFormat(v time.Time) string {
	//return v.Format(time.Kitchen)
	return v.Format("15:04")
}

func printSchedule(city string, t1, t2 prayer.Schedule) string {
	format := func(k string, v, x time.Time) string {
		return fmt.Sprintf(`<div><span id="name">%s</span><span id="time">%s / %s</span></div>`, k, dateFormat(v), dateFormat(x))
	}

	str := fmt.Sprintf(`<h2 id="%s">%s</h2>`, strings.ReplaceAll(city, " ", ""), city)
	str += fmt.Sprintf(`<div style="font-size: 0.5em; margin-bottom: 20px;"><span id="name">%s</span><span id="time">%s / %s</span></div>`, "SALAH", "TODAY", "TOMORROW")
	str += format("Fajr", t1.Fajr, t2.Fajr)
	str += format(`<span style="font-style: normal">🌅</span>`, t1.Sunrise, t2.Sunrise)
	str += format("Zuhr", t1.Zuhr, t2.Zuhr)
	str += format("Asr", t1.Asr, t2.Asr)
	str += format("Maghrib", t1.Maghrib, t2.Maghrib)
	str += format("Isha", t1.Isha, t2.Isha)
	str += "<br><br>"
	return str
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	date := time.Now().Format(time.DateOnly)

	nav := ""
	content := ""

	for _, city := range cities {
		// Calculate prayer schedule in London for 2023.
		// Since London in higher latitude, make sure to enable the adapter.
		tz, _ := time.LoadLocation(city.Location)
		schedules, _ := prayer.Calculate(prayer.Config{
			Latitude:            city.Lat,
			Longitude:           city.Lon,
			Timezone:            tz,
			TwilightConvention:  prayer.MWL(),
			AsrConvention:       prayer.Shafii,
			HighLatitudeAdapter: prayer.NearestLatitude(),
			PreciseToSeconds:    true,
		}, 2024)

		nav += fmt.Sprintf(`<a href="#%s" class="head">%s</a>`, strings.ReplaceAll(city.Name, " ", ""), city.Name)

		for i, sched := range schedules {
			if sched.Date != date {
				continue
			}

			content += printSchedule(city.Name, schedules[i], schedules[i+1])
		}
	}

	out := mu.Template("Pray", "Islamic Prayer Times", nav, fmt.Sprintf(template, content))
	mu.Render(w, out)
	return
}

func Register() {}
