package stats

import (
	"sync"
	"time"
	"strconv"
	"github.com/mssola/user_agent"
	"github.com/kataras/iris"
)

type (
	Stats struct {
		Uptime       		time.Time      `json:"uptime"`
		RequestCount 		uint64         `json:"requestCount"`
		Statuses			map[string]int `json:"statuses"`
		Methods      		map[string]int `json:"methods"`
		Paths        		map[string]int `json:"paths"`
		Platform	 		map[string]int `json:"platforms"`
		OS			 		map[string]int `json:"os"`
		BrowserName			map[string]int `json:"browsername"`
		BrowserVersion		map[string]int `json:"browserversion"`
		Mobiles			 	map[string]int `json:"mobiles"`
		Bots			 	map[string]int `json:"bots"`
		mutex        sync.RWMutex
	}
)

func (s *Stats) Serve(ctx *iris.Context)  {
	defer func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		//elapsed := time.Since(ctx.Time())
		ua := user_agent.New(string(ctx.UserAgent()))
		status := strconv.Itoa(ctx.Response.StatusCode())
		browserName, browserVersion := ua.Browser()
		plat := "API"
		if(ua.Platform() != "") {plat = ua.Platform()}
		os := "NA"
		if(ua.OS() != "") {os = ua.OS()}
		
		s.RequestCount++
		s.Statuses[status]++
		s.Methods[ctx.MethodString()]++
		/* This only works partially..../stat is added randomly with int of Nil, which then fails to render.
		if _, ok := s.Paths[ctx.PathString()]; ok {
			s.Paths[ctx.PathString()]++
		} else {
			fmt.Println("HERE")
			s.Paths[ctx.PathString()] = 1
		}
		*/
		s.Platform[plat]++
		s.OS[os]++
		s.BrowserName[browserName]++
		s.BrowserVersion[browserName + " " +browserVersion]++
		s.Mobiles[strconv.FormatBool(ua.Mobile())]++
		s.Bots[strconv.FormatBool(ua.Bot())]++
	}()
	ctx.Next()
}

func New() *Stats {
	return &Stats{
		Uptime:   			time.Now(),
		Statuses: 			make(map[string]int),
		Methods:  			make(map[string]int),
		Paths:    			make(map[string]int),
		Platform:			make(map[string]int),
		OS:			 		make(map[string]int),
		BrowserName:		make(map[string]int),
		BrowserVersion:		make(map[string]int),
		Mobiles:		 	make(map[string]int),
		Bots:			 	make(map[string]int),
	}
}

func (s *Stats) Handle(ctx *iris.Context)  {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	//ctx.JSON(iris.StatusOK, s)
	ctx.Render("application/json", s, iris.RenderOptions{"charset": "UTF-8"}) // UTF-8 is the default.
		
}
