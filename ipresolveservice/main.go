package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/kostz/ipresolveservice/cache"
)

// LISTEN defines the listening port
const LISTEN = "0.0.0.0:5501"

// IPMetadata contains metadata to be retrieved
type IPMetadata struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
}

// IPResolveService holds the service components
type IPResolveService struct {
	cache  cache.Cache
	logger *zap.Logger
}

// NewIPResolveService creates the instance of a service
func NewIPResolveService() *IPResolveService {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return &IPResolveService{
		cache:  cache.NewSyncCache(),
		logger: logger,
	}
}

func (s *IPResolveService) requestMetadata(ip string) IPMetadata {
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

// ResolveIP handles the /api/v1/resolve endpoint
func (s *IPResolveService) ResolveIP(w http.ResponseWriter, r *http.Request) {
	var (
		data      IPMetadata
		dataRaw   interface{}
		ipAddress string
		ok        bool
	)

	ipAddress = mux.Vars(r)["ip"]
	s.logger.Info("rq received", zap.String("ip", ipAddress))
	if _, ok = s.cache.Get(ipAddress); !ok {
		data = s.requestMetadata(ipAddress)
		s.cache.Set(ipAddress, data)
	}

	dataRaw, _ = s.cache.Get(ipAddress)

	data, ok = dataRaw.(IPMetadata)
	if !ok {
		s.logger.Error("can't cast data")
	}
	enc, _ := json.Marshal(data)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(enc)
	if err != nil {
		s.logger.Error("can't write response", zap.Error(err))
	}
}

func (s *IPResolveService) start() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/resolve/{ip}", s.ResolveIP).Methods(http.MethodGet)

	srv := &http.Server{}
	srv.Addr = LISTEN
	srv.Handler = r

	s.logger.Info("server starting...")
	func() {
		err := srv.ListenAndServe()
		zap.Error(err)
	}()

}

func main() {
	s := NewIPResolveService()
	s.start()
}
