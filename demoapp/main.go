package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"net/http"
	"time"
)

func main() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		ipv4 := gofakeit.IPv4Address()
		eventId := gofakeit.UUID()
		userId := gofakeit.Username()
		tz := gofakeit.TimeZoneAbv()
		data := map[string]string{"ip_addr": ipv4, "eventId": eventId,
			"userId": userId, "timezone": tz}
		mapData, _ := json.Marshal(data)
		// report log record to otelcol pipeline
		r, err := http.Post("http://0.0.0.0:5520/report", "application/json", bytes.NewBuffer(mapData))
		fmt.Println(r.StatusCode, err)
	}
}
