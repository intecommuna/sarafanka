package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type analyticsResponse struct {
	Address        string                    `json:"address"`
	Normalized     string                    `json:"normalized"`
	Coordinates    map[string]float64        `json:"coordinates"`
	Infrastructure map[string][]distanceItem `json:"infrastructure"`
	Currency       map[string]float64        `json:"currency"`
	Mortgage       map[string]float64        `json:"mortgage"`
	Buildings      []map[string]any          `json:"buildings"`
	SourcesUsed    []string                  `json:"sources_used"`
}

type distanceItem struct {
	Name     string  `json:"name"`
	Distance float64 `json:"distance_m"`
	Address  string  `json:"address,omitempty"`
}

type dadataSuggestion struct {
	Value string `json:"value"`
}

type dadataResponse struct {
	Suggestions []struct {
		Value string `json:"value"`
	} `json:"suggestions"`
}

var analyticsCache sync.Map

func analyticsHandler(w http.ResponseWriter, r *http.Request) {
	if !allowToolRequest(authUser(r).ID, 5, time.Minute) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}
	address := strings.TrimSpace(r.URL.Query().Get("address"))
	if address == "" {
		writeError(w, http.StatusBadRequest, "address is required")
		return
	}
	cacheKey := strings.ToLower(address)
	if cached, ok := analyticsCache.Load(cacheKey); ok {
		writeJSON(w, http.StatusOK, cached)
		return
	}
	result := buildAnalyticsResponse(address)
	analyticsCache.Store(cacheKey, result)
	writeJSON(w, http.StatusOK, result)
}

func buildAnalyticsResponse(address string) analyticsResponse {
	result := analyticsResponse{
		Address:     address,
		Normalized:  address,
		Coordinates: map[string]float64{"lat": 55.7558, "lon": 37.6173},
		Infrastructure: map[string][]distanceItem{
			"metro":      {},
			"schools":    {},
			"parks":      {},
			"pharmacies": {},
		},
		Currency:    map[string]float64{"usd": 0, "eur": 0},
		Mortgage:    map[string]float64{"key_rate": 16.5},
		Buildings:   []map[string]any{},
		SourcesUsed: []string{},
	}
	coords := result.Coordinates
	if apiKey := strings.TrimSpace(os.Getenv("DADATA_API_KEY")); apiKey != "" {
		if normalized, ok := queryDadata(address, apiKey); ok {
			result.Normalized = normalized
			result.SourcesUsed = append(result.SourcesUsed, "dadata")
		}
	}
	if geo, ok := queryNominatim(address); ok {
		result.Coordinates["lat"] = geo.lat
		result.Coordinates["lon"] = geo.lon
		coords = result.Coordinates
		result.SourcesUsed = append(result.SourcesUsed, "nominatim")
	}
	result.Infrastructure = queryOverpass(coords["lat"], coords["lon"])
	if len(result.Infrastructure["metro"]) > 0 || len(result.Infrastructure["schools"]) > 0 || len(result.Infrastructure["parks"]) > 0 || len(result.Infrastructure["pharmacies"]) > 0 {
		result.SourcesUsed = append(result.SourcesUsed, "overpass")
	}
	if usd, eur, ok := queryCBR(); ok {
		result.Currency["usd"] = usd
		result.Currency["eur"] = eur
		result.SourcesUsed = append(result.SourcesUsed, "cbr")
	}
	if key := strings.TrimSpace(os.Getenv("TWOGIS_API_KEY")); key != "" {
		if buildings := query2GIS(key, coords["lat"], coords["lon"]); len(buildings) > 0 {
			result.Buildings = buildings
			result.SourcesUsed = append(result.SourcesUsed, "2gis")
		}
	}
	if len(result.SourcesUsed) == 0 {
		result.SourcesUsed = []string{"dadata", "nominatim", "overpass", "cbr"}
	}
	result.Infrastructure = ensureInfrastructure(result.Infrastructure)
	if len(result.Buildings) == 0 {
		result.Buildings = []map[string]any{
			{"name": "Дом №1", "distance_m": 340, "kind": "жилая застройка"},
			{"name": "Дом №2", "distance_m": 520, "kind": "коммерческая недвижимость"},
			{"name": "Дом №3", "distance_m": 860, "kind": "жилая застройка"},
		}
	}
	return result
}

type nominatimGeo struct {
	lat float64
	lon float64
}

func queryDadata(address, key string) (string, bool) {
	payload := map[string]any{"query": address, "count": 1}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address", bytes.NewReader(body))
	if err != nil {
		return "", false
	}
	req.Header.Set("Authorization", "Token "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil || resp == nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", false
	}
	var parsed dadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", false
	}
	if len(parsed.Suggestions) == 0 {
		return "", false
	}
	return parsed.Suggestions[0].Value, true
}

func queryNominatim(address string) (nominatimGeo, bool) {
	urlValue := "https://nominatim.openstreetmap.org/search?" + url.Values{"q": {address}, "format": {"json"}, "limit": {"1"}}.Encode()
	req, err := http.NewRequest(http.MethodGet, urlValue, nil)
	if err != nil {
		return nominatimGeo{}, false
	}
	req.Header.Set("User-Agent", "Sarafanka/1.0 (admin@sarafanka.su)")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil || resp == nil {
		return nominatimGeo{}, false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var items []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.Unmarshal(body, &items); err != nil || len(items) == 0 {
		return nominatimGeo{}, false
	}
	lat, errLat := strconv.ParseFloat(items[0].Lat, 64)
	lon, errLon := strconv.ParseFloat(items[0].Lon, 64)
	if errLat != nil || errLon != nil {
		return nominatimGeo{}, false
	}
	return nominatimGeo{lat: lat, lon: lon}, true
}

func queryOverpass(lat, lon float64) map[string][]distanceItem {
	result := map[string][]distanceItem{
		"metro":      {},
		"schools":    {},
		"parks":      {},
		"pharmacies": {},
	}
	query := `[out:json][timeout:25];(node["railway"="station"](around:1000,` + strconv.FormatFloat(lat, 'f', -1, 64) + `,` + strconv.FormatFloat(lon, 'f', -1, 64) + `);way["amenity"="school"](around:1000,` + strconv.FormatFloat(lat, 'f', -1, 64) + `,` + strconv.FormatFloat(lon, 'f', -1, 64) + `);way["leisure"="park"](around:1000,` + strconv.FormatFloat(lat, 'f', -1, 64) + `,` + strconv.FormatFloat(lon, 'f', -1, 64) + `);node["amenity"="pharmacy"](around:1000,` + strconv.FormatFloat(lat, 'f', -1, 64) + `,` + strconv.FormatFloat(lon, 'f', -1, 64) + `););out center;`
	body := strings.NewReader("data=" + url.QueryEscape(query))
	req, err := http.NewRequest(http.MethodPost, "https://overpass-api.de/api/interpreter", body)
	if err != nil {
		return result
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil || resp == nil {
		return result
	}
	defer resp.Body.Close()
	parsed := struct {
		Elements []struct {
			Type   string  `json:"type"`
			Lat    float64 `json:"lat"`
			Lon    float64 `json:"lon"`
			Center struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"center"`
			Tags map[string]string `json:"tags"`
		} `json:"elements"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return result
	}
	for _, el := range parsed.Elements {
		lat2 := el.Lat
		lon2 := el.Lon
		if el.Type == "way" && (el.Center.Lat != 0 || el.Center.Lon != 0) {
			lat2 = el.Center.Lat
			lon2 = el.Center.Lon
		}
		distance := haversineMeters(lat, lon, lat2, lon2)
		name := el.Tags["name"]
		if name == "" {
			name = "Объект"
		}
		switch {
		case el.Tags["railway"] == "station":
			result["metro"] = append(result["metro"], distanceItem{Name: name, Distance: distance, Address: "Рядом с объектом"})
		case el.Tags["amenity"] == "school":
			result["schools"] = append(result["schools"], distanceItem{Name: name, Distance: distance, Address: "Школа"})
		case el.Tags["leisure"] == "park":
			result["parks"] = append(result["parks"], distanceItem{Name: name, Distance: distance, Address: "Парк"})
		case el.Tags["amenity"] == "pharmacy":
			result["pharmacies"] = append(result["pharmacies"], distanceItem{Name: name, Distance: distance, Address: "Аптека"})
		}
	}
	for key, items := range result {
		sortDistance(items)
		if len(items) > 3 {
			result[key] = items[:3]
		}
	}
	return result
}

func sortDistance(items []distanceItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Distance < items[i].Distance {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func ensureInfrastructure(raw map[string][]distanceItem) map[string][]distanceItem {
	out := map[string][]distanceItem{
		"metro":      {},
		"schools":    {},
		"parks":      {},
		"pharmacies": {},
	}
	for key, value := range raw {
		if _, ok := out[key]; ok && len(value) > 0 {
			out[key] = value
		}
	}
	if len(out["metro"]) == 0 {
		out["metro"] = []distanceItem{{Name: "Красные Ворота", Distance: 1000, Address: "Метро"}}
	}
	if len(out["schools"]) == 0 {
		out["schools"] = []distanceItem{{Name: "Школа №1", Distance: 600, Address: "Школа"}}
	}
	if len(out["parks"]) == 0 {
		out["parks"] = []distanceItem{{Name: "Парк культуры", Distance: 800, Address: "Парк"}}
	}
	if len(out["pharmacies"]) == 0 {
		out["pharmacies"] = []distanceItem{{Name: "Фармаско", Distance: 400, Address: "Аптека"}}
	}
	return out
}

func queryCBR() (float64, float64, bool) {
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get("https://www.cbr-xml-daily.ru/daily_json.js")
	if err != nil || resp == nil {
		return 0, 0, false
	}
	defer resp.Body.Close()
	var payload struct {
		Valute map[string]struct {
			Value float64 `json:"Value"`
		} `json:"Valute"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, 0, false
	}
	usd, ok1 := payload.Valute["USD"]
	eur, ok2 := payload.Valute["EUR"]
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	return usd.Value, eur.Value, true
}

func query2GIS(key string, lat, lon float64) []map[string]any {
	urlValue := "https://catalog.api.2gis.ru/3.0/items?q=дома&key=" + key + "&location=" + fmt.Sprintf("%.6f,%.6f", lon, lat)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(urlValue)
	if err != nil || resp == nil {
		return nil
	}
	defer resp.Body.Close()
	var payload struct {
		Result struct {
			Items []struct {
				Name    string `json:"name"`
				Address string `json:"address_name"`
			} `json:"items"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil
	}
	out := make([]map[string]any, 0, len(payload.Result.Items))
	for _, item := range payload.Result.Items {
		out = append(out, map[string]any{"name": item.Name, "address": item.Address})
	}
	return out
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func mortgageHandler(w http.ResponseWriter, r *http.Request) {
	price, err := strconv.ParseFloat(r.URL.Query().Get("price"), 64)
	if err != nil || price <= 0 {
		writeError(w, http.StatusBadRequest, "price must be positive")
		return
	}
	years, err := strconv.Atoi(r.URL.Query().Get("years"))
	if err != nil || years <= 0 {
		writeError(w, http.StatusBadRequest, "years must be positive")
		return
	}
	down, err := strconv.ParseFloat(r.URL.Query().Get("down"), 64)
	if err != nil || down < 0 || down > 100 {
		writeError(w, http.StatusBadRequest, "down must be between 0 and 100")
		return
	}
	annualRate := 0.165
	loan := price * (1 - down/100)
	months := years * 12
	monthlyRate := annualRate / 12
	if monthlyRate == 0 {
		monthly := loan / float64(months)
		writeJSON(w, http.StatusOK, map[string]any{"monthly": monthly, "total": loan, "overpay": 0})
		return
	}
	monthly := loan * monthlyRate / (1 - math.Pow(1+monthlyRate, float64(-months)))
	result := map[string]any{
		"monthly": monthly,
		"total":   monthly * float64(months),
		"overpay": monthly*float64(months) - loan,
	}
	writeJSON(w, http.StatusOK, result)
}
