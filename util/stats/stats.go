// MIT License
//
// Copyright (c) 2020 The vine Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package stats

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type stats struct {
	sync.RWMutex

	Started int64  `json:"started"`
	Memory  string `json:"memory"`
	Threads int    `json:"threads"`
	GC      string `json:"gc_pause"`

	Counters []*counter `json:"counters"`

	running bool
	exit    chan bool
}

type counter struct {
	// time created
	Timestamp int64 `json:"timestamp"`
	// counters
	Status map[string]int `json:"status_codes"`
	Total  int            `json:"total_reqs"`
}

var (
	// 5 second window
	window = time.Second * 5
	// 120 seconds total
	total = 24
)

func render(ctx *gin.Context, tmpl string, data interface{}) {
	t, err := template.New("template").Funcs(template.FuncMap{
		//		"format": format,
	}).Parse(layoutTemplate)
	if err != nil {
		ctx.JSON(500, "Error occurred:"+err.Error())
		return
	}
	t, err = t.Parse(tmpl)
	if err != nil {
		ctx.JSON(500, "Error occurred:"+err.Error())
		return
	}
	if err = t.ExecuteTemplate(ctx.Writer, "layout", data); err != nil {
		ctx.JSON(500, "Error occurred:"+err.Error())
		return
	}
}

func (s *stats) run() {
	t := time.NewTicker(window)
	w := 0

	for {
		select {
		case <-s.exit:
			t.Stop()
			return
		case <-t.C:
			// roll
			s.Lock()
			s.Counters = append(s.Counters, &counter{
				Timestamp: time.Now().Unix(),
				Status:    make(map[string]int),
			})
			if len(s.Counters) >= total {
				s.Counters = s.Counters[1:]
			}

			w++
			if w >= 2 {
				var mstat runtime.MemStats
				runtime.ReadMemStats(&mstat)
				s.Threads = runtime.NumGoroutine()
				s.Memory = fmt.Sprintf("%.2fmb", float64(mstat.Alloc)/float64(1024*1024))
				s.GC = fmt.Sprintf("%.3fms", float64(mstat.PauseTotalNs)/(1000*1000))
				w = 0
			}
			s.Unlock()
		}
	}
}

func (s *stats) Record(c string, t int) {
	s.Lock()
	counter := s.Counters[len(s.Counters)-1]
	counter.Status[c] += t
	counter.Total += t
	s.Counters[len(s.Counters)-1] = counter
	s.Unlock()
}

func (s *stats) ServeHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var code string
		rw := &writer{w, 200}

		h.ServeHTTP(w, r)

		switch {
		case rw.status >= 500:
			code = "50x"
		case rw.status >= 400:
			code = "40x"
		case rw.status >= 300:
			code = "30x"
		case rw.status >= 200:
			code = "20x"
		}

		s.Record(code, 1)
	})
}

func (s *stats) StatsHandler(ctx *gin.Context) {
	if ct := ctx.GetHeader("Content-Type"); ct == "application/json" {
		s.RLock()
		b, err := json.Marshal(s)
		s.RUnlock()
		if err != nil {
			ctx.JSON(500, err.Error())
			return
		}
		ctx.JSON(200, b)
		return
	}

	render(ctx, statsTemplate, nil)
}

func (s *stats) Start() error {
	s.Lock()
	defer s.Unlock()

	if s.running {
		return nil
	}

	s.Started = time.Now().Unix()
	s.exit = make(chan bool)
	go s.run()
	return nil
}

func (s *stats) Stop() error {
	s.Lock()
	defer s.Unlock()

	if !s.running {
		return nil
	}

	close(s.exit)
	s.Started = 0
	return nil
}

func New() *stats {
	var mstat runtime.MemStats
	runtime.ReadMemStats(&mstat)

	return &stats{
		Threads: runtime.NumGoroutine(),
		Memory:  fmt.Sprintf("%.2fmb", float64(mstat.Alloc)/float64(1024*1024)),
		GC:      fmt.Sprintf("%.3fms", float64(mstat.PauseTotalNs)/(1000*1000)),
		Counters: []*counter{
			&counter{
				Timestamp: time.Now().Unix(),
				Status:    make(map[string]int),
			},
		},
	}
}
