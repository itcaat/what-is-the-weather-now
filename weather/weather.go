package weather

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/itcaat/what-is-the-weather-now/cache"
)

var weatherCache = cache.NewCache()

func GetWeather(city string) (string, bool, time.Duration) {
	if cached, found, ttl := weatherCache.Get(city); found {
		return cached, true, ttl
	}

	url := fmt.Sprintf("http://wttr.in/%s?format=%%C+%%t&lang=en", city)
	resp, err := http.Get(url)
	if err != nil {
		return "", false, 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false, 0
	}

	weather := string(body)
	ttl := 10 * time.Minute

	weatherCache.Set(city, weather, ttl)

	return weather, false, ttl
}
