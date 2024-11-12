package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"

	"github.com/kostz/ipresolveservice/cache"
)

const LISTEN = "0.0.0.0:5501"

type IPMetadata struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
}

type IpResolveService struct {
	cache cache.Cache
}

func NewIpResolveService() *IpResolveService {
	return &IpResolveService{
		cache: cache.NewSyncCache(),
	}
}

func (s *IpResolveService) requestMetadata(ip string) IPMetadata {
	res := IPMetadata{}
	resp, err := http.Get(
		fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,city,region,timezone,currency", ip),
	)

	if err != nil {
		return res
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return res
	}

	err = json.Unmarshal(respData, &res)
	if err != nil {
		return res
	}

	return res
}

func (s *IpResolveService) ResolveIP(w http.ResponseWriter, r *http.Request) {
	var (
		data      IPMetadata
		dataRaw   interface{}
		ipAddress string
		ok        bool
	)

	ipAddress = mux.Vars(r)["ip"]
	if _, ok = s.cache.Get(ipAddress); !ok {
		data = s.requestMetadata(ipAddress)
		s.cache.Set(ipAddress, data)
	}

	dataRaw, ok = s.cache.Get(ipAddress)
	data, ok = dataRaw.(IPMetadata)

	enc, _ := json.Marshal(data)
	w.WriteHeader(http.StatusOK)
	w.Write(enc)
}

func (s *IpResolveService) start() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/resolve/{ip}", s.ResolveIP).Methods(http.MethodGet)

	srv := &http.Server{}
	srv.Addr = LISTEN
	srv.Handler = r

	func() {
		err := srv.ListenAndServe()
		panic(err)
	}()
}

func main() {
	s := NewIpResolveService()
	s.start()
}
