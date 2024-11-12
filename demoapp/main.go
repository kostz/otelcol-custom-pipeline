package main

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		ipv4 := gofakeit.IPv4Address()
		eventId := gofakeit.UUID()
		userId := gofakeit.Username()
		tz := gofakeit.TimeZoneAbv()
		data := map[string]string{"ip_addr": ipv4, "eventId": eventId,
			"userId": userId, "timezone": tz}
		mapData, _ := json.Marshal(data)
		fmt.Println(string(mapData))
	}
}
