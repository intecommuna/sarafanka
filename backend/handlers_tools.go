package main

import (
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type parserListing struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
	Rooms    int    `json:"rooms"`
	Price    int    `json:"price"`
	Source   string `json:"source"`
	URL      string `json:"url"`
}

var (
	parserSeedOnce sync.Once
	parserSeedData []parserListing
	toolRateMu     sync.Mutex
	toolBuckets    = map[int64][]time.Time{}
)

func parserDataset() []parserListing {
	parserSeedOnce.Do(func() {
		cities := []string{"Москва", "Москва", "Москва", "Санкт-Петербург", "Казань", "Екатеринбург"}
		districts := []string{"Центр", "Митино", "Тушино", "Хамовники", "Южное Бутово", "Перово", "Красная Пресня", "Крылатское", "Видное", "Бибирево"}
		sources := []struct {
			name   string
			prefix string
		}{
			{name: "avito", prefix: "Авито"},
			{name: "cian", prefix: "Циан"},
			{name: "domclick", prefix: "Домклик"},
		}
		var seed []parserListing
		for i := 0; i < 30; i++ {
			src := sources[i%len(sources)]
			city := cities[i%len(cities)]
			district := districts[i%len(districts)]
			rooms := 1 + (i % 4)
			basePrice := 2400000 + i*170000 + (i%5)*90000
			if rooms >= 3 {
				basePrice += 1500000
			}
			seed = append(seed, parserListing{
				ID:       1000 + i,
				Title:    src.prefix + " · " + strconv.Itoa(rooms) + "-комнатная квартира",
				City:     city,
				District: district,
				Address:  city + ", " + district + ", ул. " + strconv.Itoa(10+i) + ", д. " + strconv.Itoa(1+i),
				Rooms:    rooms,
				Price:    basePrice,
				Source:   src.name,
				URL:      "https://example.com/" + src.name + "/" + strconv.Itoa(1000+i),
			})
		}
		parserSeedData = seed
	})
	return append([]parserListing(nil), parserSeedData...)
}

func allowToolRequest(uid int64, limit int, window time.Duration) bool {
	toolRateMu.Lock()
	defer toolRateMu.Unlock()
	cutoff := time.Now().Add(-window)
	alive := toolBuckets[uid][:0]
	for _, ts := range toolBuckets[uid] {
		if ts.After(cutoff) {
			alive = append(alive, ts)
		}
	}
	toolBuckets[uid] = alive
	if len(alive) >= limit {
		return false
	}
	toolBuckets[uid] = append(alive, time.Now())
	return true
}

func parserHandler(w http.ResponseWriter, r *http.Request) {
	if !allowToolRequest(authUser(r).ID, 10, time.Minute) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}
	query := r.URL.Query()
	city := strings.TrimSpace(strings.ToLower(query.Get("city")))
	district := strings.TrimSpace(strings.ToLower(query.Get("district")))
	source := strings.TrimSpace(strings.ToLower(query.Get("source")))
	roomsRaw := strings.TrimSpace(query.Get("rooms"))
	priceMinRaw := strings.TrimSpace(query.Get("price_min"))
	priceMaxRaw := strings.TrimSpace(query.Get("price_max"))
	rooms, _ := strconv.Atoi(roomsRaw)
	priceMin, _ := strconv.Atoi(priceMinRaw)
	priceMax, _ := strconv.Atoi(priceMaxRaw)
	if rooms < 0 {
		rooms = 0
	}
	items := make([]parserListing, 0)
	for _, item := range parserDataset() {
		matchCity := city == "" || strings.Contains(strings.ToLower(item.City), city)
		matchDistrict := district == "" || strings.Contains(strings.ToLower(item.District), district)
		matchRooms := rooms == 0 || item.Rooms == rooms
		matchSource := source == "" || source == "all" || strings.EqualFold(item.Source, source)
		matchMin := priceMin <= 0 || item.Price >= priceMin
		matchMax := priceMax <= 0 || item.Price <= priceMax
		if matchCity && matchDistrict && matchRooms && matchSource && matchMin && matchMax {
			adjusted := item
			adjusted.Price = int(float64(adjusted.Price) * (1 + (rand.Float64()-0.5)*0.1))
			items = append(items, adjusted)
		}
	}
	if sortKey := strings.TrimSpace(query.Get("sort")); sortKey != "" {
		switch sortKey {
		case "price_asc":
			sort.Slice(items, func(i, j int) bool { return items[i].Price < items[j].Price })
		case "price_desc":
			sort.Slice(items, func(i, j int) bool { return items[i].Price > items[j].Price })
		case "rooms_desc":
			sort.Slice(items, func(i, j int) bool { return items[i].Rooms > items[j].Rooms })
		default:
			sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
		}
	} else {
		sort.Slice(items, func(i, j int) bool { return items[i].Price < items[j].Price })
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}
