package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type IPInfo struct {
	City string `json:"city"`
}

func getIPInfo(ip string) (string, bool) {
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "Moscow", false // Если не удалось определить, используем Москву
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "Moscow", false
	}

	if info.City == "" {
		return "Moscow", false
	}
	return info.City, true
}

func getWeather(city string) (string, error) {
	url := fmt.Sprintf("http://wttr.in/%s?format=%%C+%%t&lang=ru", city)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Get client IP from Cloudflare headers
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0] // Fallback to RemoteAddr if header is missing
	}

	city, detected := getIPInfo(ip)

	weather, err := getWeather(city)
	if err != nil {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}

	message := ""
	if !detected {
		message = "(Location could not be determined, using Moscow as default)"
	}

	html := fmt.Sprintf(`
        <html>
        <head><title>Weather</title><meta charset="UTF-8"></head>
        <body>
            <h1>Your IP: %s</h1>
            <h2>Weather in %s %s</h2>
            <p>%s</p>
        </body>
        </html>
    `, ip, city, message, weather)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Lister port:", port)
	http.ListenAndServe(":"+port, nil)
}
