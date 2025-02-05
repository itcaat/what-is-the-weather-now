package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/itcaat/what-is-the-weather-now/cache"
	"github.com/itcaat/what-is-the-weather-now/weather"
)

var ipCache = cache.NewCache()

func getIPInfo(ip string) (string, bool) {
	if cached, found, _ := ipCache.Get(ip); found {
		return cached, true
	}

	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "Moscow", false
	}
	defer resp.Body.Close()

	var result struct {
		City string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.City == "" {
		return "Moscow", false
	}

	ipCache.Set(ip, result.City, 24*time.Hour)
	return result.City, true
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	city, detected := getIPInfo(ip)
	weatherData, cached, ttl := weather.GetWeather(city)

	message := ""
	if !detected {
		message = "(Location could not be determined, using Moscow as default)"
	}

	cacheInfo := ""
	if cached {
		cacheInfo = fmt.Sprintf("(Cached data, TTL remaining: %v)", ttl.Round(time.Second))
	}

	html := fmt.Sprintf(`
		<html>
		<head><title>Weather</title><meta charset="UTF-8"></head>
		<body>
			<h1>Your IP: %s</h1>
			<h2>Weather in %s %s</h2>
			<p>%s %s</p>
		</body>
		</html>
	`, ip, city, message, weatherData, cacheInfo)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
