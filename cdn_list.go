package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"
)

type resolutions struct {
	IPAdddress string `json:"ip_address"`
	LastResolved string `json:"last_resolved"`
}
type threatCrowdAPIResult struct {
	Resolutions []resolutions `json:"resolutions"`
}

func getAddresses() (addr []net.IP) {
	hres, err := http.Get("https://www.threatcrowd.org/searchApi/v2/domain/report/?domain=pbs.twimg.com")
	if err != nil {
		panic(err)
	}
	defer hres.Body.Close()

	var res threatCrowdAPIResult
	err = json.NewDecoder(hres.Body).Decode(&res)
	if err != nil {
		panic(err)
	}

	// 1년
	minDate := time.Now().Add((time.Duration)(-365 * 24) * time.Hour)

	for _, resolution := range res.Resolutions {
		lastResolved, err := time.Parse("2006-01-02", resolution.LastResolved)
		if err != nil {
			continue
		}

		if lastResolved.Before(minDate) {
			continue
		}

		ip := net.ParseIP(resolution.IPAdddress)
		if ip.To4() != nil {
			addr = append(addr, ip)
		}
	}

	if len(addr) == 0 {
		panic("cdn 목록을 가져올 수 없었습니다.")
	}

	return
}