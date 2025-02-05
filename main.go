package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

type IPInfo struct {
	City string `json:"city"`
}

var weatherCache = cache.New(10*time.Minute, 15*time.Minute)
var ipCache = cache.New(24*time.Hour, 30*time.Minute)

func getIPInfo(ip string) (string, bool) {
	// Check if IP is already cached
	if cachedCity, found := ipCache.Get(ip); found {
		return cachedCity.(string), true
	}

	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "Moscow", false // Default to Moscow if location cannot be determined
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "Moscow", false
	}

	if info.City == "" {
		return "Moscow", false
	}

	// Store in cache
	ipCache.Set(ip, info.City, cache.DefaultExpiration)

	return info.City, true
}

func getWeather(city string) (string, bool, time.Duration) {
	// Check if weather data is cached
	if cachedWeather, found := weatherCache.Get(city); found {
		item := cachedWeather.(struct {
			weather string
			expiry  time.Time
		})
		ttl := time.Until(item.expiry)
		return item.weather, true, ttl
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

	// Store in cache
	weatherCache.Set(city, struct {
		weather string
		expiry  time.Time
	}{weather, time.Now().Add(ttl)}, cache.DefaultExpiration)

	return weather, false, ttl
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Determine client IP, checking Cloudflare headers first
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0] // Fallback to RemoteAddr if header is missing
	}

	city, detected := getIPInfo(ip)

	weather, cached, ttl := getWeather(city)
	if weather == "" {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}

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
	`, ip, city, message, weather, cacheInfo)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server started on port:", port)
	http.ListenAndServe(":"+port, nil)
}
