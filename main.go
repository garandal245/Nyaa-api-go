package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var logIPs bool

func sendJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control, Content-Type")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, status int, msg string) {
	sendJSON(w, status, map[string]string{"error": msg})
}

func getQuery(r *http.Request, key, def string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	return v
}

func sortParam(r *http.Request) string {
	// Nyaa uses "id" internally but the API accepts "date" as an alias
	s := r.URL.Query().Get("s")
	if s == "date" {
		return "id"
	}
	return s
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if logIPs {
			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
				if i := strings.LastIndex(ip, ":"); i != -1 {
					ip = ip[:i]
				}
				ip = strings.Trim(ip, "[]")
			}
			log.Printf("%s %s — %s", r.Method, r.URL.RequestURI(), ip)
		} else {
			log.Printf("%s %s", r.Method, r.URL.RequestURI())
		}
		next(w, r)
	}
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	sendJSON(w, 200, "Nyaa API v2 // Alive")
}

func handleID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/id/")
	if id == "" {
		sendError(w, 400, "missing id")
		return
	}

	file, err := scrapeDetail(NyaaBaseURL + "/view/" + id)
	if err != nil {
		log.Printf("scrapeDetail error: %v", err)
		sendError(w, 404, "Not Found")
		return
	}
	sendJSON(w, 200, file)
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/user/")
	if username == "" {
		sendError(w, 400, "missing username")
		return
	}

	params := QueryParams{
		Query:  r.URL.Query().Get("q"),
		Page:   getQuery(r, "p", "1"),
		Sort:   sortParam(r),
		Order:  getQuery(r, "o", "desc"),
		Filter: getQuery(r, "f", "0"),
	}

	q := strings.ReplaceAll(params.Query, " ", "+")
	url := fmt.Sprintf("%s/user/%s?q=%s&p=%s&s=%s&o=%s&f=%s",
		NyaaBaseURL, username, q, params.Page, params.Sort, params.Order, params.Filter)

	torrents, err := scrapeList(url)
	if err != nil {
		log.Printf("scrapeList error: %v", err)
		sendError(w, 404, "Not Found")
		return
	}
	sendJSON(w, 200, torrents)
}

func handleCategory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)

	cat := parts[0]
	sub := "all"
	if len(parts) == 2 && parts[1] != "" {
		sub = parts[1]
	}

	subMap, ok := categoryMap[cat]
	if !ok {
		sendError(w, 400, fmt.Sprintf("unknown category: %s", cat))
		return
	}
	catID, ok := subMap[sub]
	if !ok {
		sendError(w, 400, fmt.Sprintf("unknown subcategory: %s", sub))
		return
	}

	params := QueryParams{
		Query:  r.URL.Query().Get("q"),
		Page:   getQuery(r, "p", "1"),
		Sort:   getQuery(r, "s", ""),
		Order:  getQuery(r, "o", "desc"),
		Filter: getQuery(r, "f", "0"),
	}

	q := strings.ReplaceAll(params.Query, " ", "+")
	url := fmt.Sprintf("%s?q=%s&c=%s&p=%s&s=%s&o=%s&f=%s",
		NyaaBaseURL, q, catID, params.Page, params.Sort, params.Order, params.Filter)

	torrents, err := scrapeList(url)
	if err != nil {
		log.Printf("scrapeList error: %v", err)
		sendError(w, 404, "Not Found")
		return
	}
	sendJSON(w, 200, torrents)
}

func router(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Cache-Control, Content-Type")
		w.WriteHeader(204)
		return
	}
	if r.Method != http.MethodGet {
		sendError(w, 405, "method not allowed")
		return
	}

	path := r.URL.Path
	switch {
	case path == "/":
		handlePing(w, r)
	case strings.HasPrefix(path, "/id/"):
		handleID(w, r)
	case strings.HasPrefix(path, "/user/"):
		handleUser(w, r)
	default:
		handleCategory(w, r)
	}
}

func main() {
	flag.BoolVar(&logIPs, "log-ips", false, "include client IP addresses in request logs")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if logIPs {
		log.Printf("Nyaa API listening on :%s (IP logging enabled)", port)
	} else {
		log.Printf("Nyaa API listening on :%s", port)
	}

	http.HandleFunc("/", withLogging(router))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
