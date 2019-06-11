package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
	
	"gopkg.in/cheggaaa/pb.v1"
)

type cdnTestResult struct {
	Addr		net.IP
	Failed		bool
	TotalTime	uint64
}

func getBestCdn(addrs []net.IP) string {
	testResult := make([]cdnTestResult, len(addrs))
	for index, addr := range addrs {
		testResult[index].Addr = addr
	}

	bar := pb.New(len(testResult))
	bar.SetMaxWidth(80)

	bar.ShowElapsedTime = true
	bar.ShowFinalTime	= false
	bar.ShowSpeed		= false
	bar.ShowTimeLeft	= false
	
	bar.Set(0)
	bar.Start()

	var g sync.WaitGroup
	g.Add(len(testResult))
	for index := range testResult {
		go cdnTest(&g, index, &testResult[index], bar)
	}

	g.Wait()

	bar.Finish()

	i := 0
	for i < len(testResult) {
		if testResult[i].Failed {
			copy(testResult[i:], testResult[i + 1:])
			testResult = testResult[:len(testResult) - 1]
		} else {
			i++
		}
	}

	sort.Slice(testResult, func (a, b int) bool { return testResult[a].TotalTime < testResult[b].TotalTime })

	for _, r := range testResult {
		fmt.Printf("%15s / %7.2d ms\n", r.Addr.String(), r.TotalTime)
	}

	if len(testResult) == 0 {
		return ""
	}
	return testResult[0].Addr.String()
}

func cdnTest(g *sync.WaitGroup, index int, r *cdnTestResult, bar *pb.ProgressBar) {
	defer func() {
		bar.Increment()
		g.Done()
	}()
	
	r.Failed = true
	
	client := http.Client {
		Timeout   : httpTimeout * time.Second,
		Transport : &http.Transport {
			Dial				: func(network, addr string) (net.Conn, error) { return net.Dial(network, strings.ReplaceAll(addr, twimgHostName, r.Addr.String())) },
			IdleConnTimeout		: httpTimeout * time.Second,
			DisableKeepAlives	: true,
		},
	}

	buff := make([]byte, httpBufferSize)
	for i := 0; i < httpCount; i++ {
		hreq, err := http.NewRequest("GET", twimgTestURI, nil)
		if err != nil {
			return
		}
		hreq.Close = true

		hres, err := client.Do(hreq)
		if err != nil {
			return
		}
		defer hres.Body.Close()

		if !strings.HasPrefix(hres.Header.Get("content-type"), "image") {
			return
		}

		start := time.Now()
		for {
			read, err := hres.Body.Read(buff)
			if err != nil && err != io.EOF {
				return
			}
			if read == 0 {
				break
			}
		}

		r.TotalTime += uint64(time.Duration(time.Now().Sub(start).Nanoseconds()) / time.Millisecond)
	}

	r.Failed = false
}