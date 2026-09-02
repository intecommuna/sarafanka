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
	ID        int    `json:"id"`
	Title     string `json:"title"`
	City      string `json:"city"`
	District  string `json:"district"`
	Address   string `json:"address"`
	Rooms     int    `json:"rooms"`
	Price     int    `json:"price"`
	Source    string `json:"source"`
	SearchURL string `json:"search_url,omitempty"`
}

var (
	parserSeedOnce sync.Once
	parserSeedData []parserListing
	toolRateMu     sync.Mutex
	toolBuckets    = map[int64][]time.Time{}
)

func parserDataset() []parserListing {
	parserSeedOnce.Do(func() {
		cities := []string{"Москва", "Санкт-Петербург", "Казань", "Екатеринбург", "Новосибирск", "Махачкала", "Южно-Сухокумск"}
		districtByCity := map[string][]string{
			"Москва":          {"Центр", "Митино", "Тушино", "Хамовники", "Южное Бутово", "Перово", "Красная Пресня", "Крылатское", "Видное", "Бибирево"},
			"Санкт-Петербург": {"Центральный", "Петроградский", "Московский", "Калининский", "Выборгский", "Невский", "Адмиралтейский", "Красногвардейский", "Фрунзенский", "Приморский"},
			"Казань":          {"Центр", "Ново-Савиновский", "Московский", "Кировский", "Приволжский", "Советский", "Авиастроительный", "Вахитовский", "Нижнекамский", "Тукая"},
			"Екатеринбург":    {"Центр", "Железнодорожный", "Орджоникидзевский", "Кировский", "Октябрьский", "Чкаловский", "Ленинский", "Верх-Исетский", "Парковый", "Академический"},
			"Новосибирск":     {"Центральный", "Заельцовский", "Кировский", "Октябрьский", "Ленинский", "Советский", "Дзержинский", "Железнодорожный", "Черепановский", "Калининский"},
			"Махачкала":       {"Центр", "Кировский", "Тарки", "Каспийский", "Южный", "Северный", "Ленинский", "Нагорный", "Порт", "Артёмовский"},
			"Южно-Сухокумск":  {"Центр", "Северный", "Южный", "Набережный", "Школьный", "Лесной", "Речной", "Промышленный", "Комсомольский", "Молодёжный"},
		}
		sources := []struct {
			name   string
			prefix string
		}{
			{name: "avito", prefix: "Авито"},
			{name: "cian", prefix: "Циан"},
			{name: "domclick", prefix: "Домклик"},
		}
		var seed []parserListing
		for _, city := range cities {
			districts := districtByCity[city]
			for i := 0; i < 10; i++ {
				src := sources[i%len(sources)]
				district := districts[i%len(districts)]
				rooms := 1 + (i % 4)
				basePrice := 1800000 + i*120000 + (i%5)*70000
				if city == "Москва" || city == "Санкт-Петербург" {
					basePrice += 1500000
				}
				if city == "Новосибирск" || city == "Екатеринбург" {
					basePrice -= 200000
				}
				if city == "Махачкала" || city == "Южно-Сухокумск" {
					basePrice = int(float64(basePrice) * 0.6)
				}
				if rooms >= 3 {
					basePrice += 500000
				}
				id := (len(seed) + 1000)
				addr := city + ", " + district + ", ул. " + strconv.Itoa(10+i) + ", д. " + strconv.Itoa(1+i)
				seed = append(seed, parserListing{
					ID:        id,
					Title:     src.prefix + " · " + strconv.Itoa(rooms) + "-комнатная квартира",
					City:      city,
					District:  district,
					Address:   addr,
					Rooms:     rooms,
					Price:     basePrice,
					Source:    src.name,
					SearchURL: sourceSearchURL(src.name, city, rooms),
				})
			}
		}
		parserSeedData = seed
	})
	return append([]parserListing(nil), parserSeedData...)
}

func sourceSearchURL(source, city string, rooms int) string {
	citySlug := normalizeCitySlug(city)
	roomLabel := "kvaritiry"
	switch rooms {
	case 1:
		roomLabel = "1-komnatnye"
	case 2:
		roomLabel = "2-komnatnye"
	case 3:
		roomLabel = "3-komnatnye"
	case 4:
		roomLabel = "4-komnatnye"
	default:
		roomLabel = "kvartiry"
	}
	switch source {
	case "avito":
		if citySlug == "" {
			return "https://www.avito.ru/rossiya/kvartiry/prodam"
		}
		if rooms > 0 && rooms <= 4 {
			return "https://www.avito.ru/" + citySlug + "/kvartiry/prodam-" + roomLabel
		}
		return "https://www.avito.ru/" + citySlug + "/kvartiry/prodam"
	case "cian":
		if citySlug == "" {
			return "https://www.cian.ru/sale/flat/"
		}
		return "https://www.cian.ru/sale/flat/"
	case "domclick":
		return "https://www.domclick.ru/"
	default:
		return "https://www.avito.ru/rossiya/kvartiry/prodam"
	}
}

func normalizeCitySlug(city string) string {
	switch strings.TrimSpace(city) {
	case "Москва":
		return "moskva"
	case "Санкт-Петербург":
		return "sankt-peterburg"
	case "Казань":
		return "kazan"
	case "Екатеринбург":
		return "ekaterinburg"
	case "Новосибирск":
		return "novosibirsk"
	case "Махачкала":
		return "mahachkala"
	case "Южно-Сухокумск":
		return "yuzhno-sukhokumsk"
	default:
		return strings.ToLower(city)
	}
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
		matchCity := city == "" || strings.EqualFold(item.City, city)
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
