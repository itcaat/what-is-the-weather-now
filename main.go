package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

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
	city := "Moscow" // По умолчанию Москва
	weather, err := getWeather(city)
	if err != nil {
		http.Error(w, "Не удалось получить погоду", http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`
		<html>
		<head><title>Погода</title><meta charset="UTF-8"></head>
		<body>
			<h1>Погода в %s</h1>
			<p>%s</p>
		</body>
		</html>
	`, city, weather)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Сервер запущен на порту:", port)
	http.ListenAndServe(":"+port, nil)
}
