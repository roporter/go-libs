package logger

import (
	"strconv"
	"time"
	"strings"
	"sync"

	"github.com/kataras/iris"
	"github.com/iris-contrib/logger"
	"github.com/reconquest/loreley"
	"github.com/fatih/color"
)

type loggerMiddleware struct {
	*logger.Logger
	config Config
}

var mapLock       sync.RWMutex

// Serve serves the middleware
func (l *loggerMiddleware) Serve(ctx *iris.Context) {
	//all except latency to string
	var date, timed, latency, status, ip, method, path string
	var startTime, endTime time.Time
	path = ctx.PathString()
	method = ctx.MethodString()

	startTime = time.Now()

	ctx.Next()
	//no time.Since in order to format it well after
	endTime = time.Now()
	timed = rightPad2Len(endTime.Format("15:04:05.999999"), "0", 15)
	date = endTime.Format("02/01/2006")
	latency = endTime.Sub(startTime).String()
	parts := strings.Split(latency,".")
	if(len(parts) == 2) {
	    if(!strings.Contains(parts[1],"ms")) {
			parts[1] = leftPad2Len(parts[1],"0",6)
		}
		latency = parts[0] + "." + parts[1]
		latency = leftPad2Len(latency," ",15)		
	} else {
		latency = leftPad2Len(latency," ",14)
	}
	
	if l.config.Status {
		status = strconv.Itoa(ctx.Response.StatusCode())
	}

	if l.config.IP {
		ip = leftPad2Len(ctx.RemoteAddr()," ",15)
	}

	if !l.config.Method {
		method = ""
	}

	if !l.config.Path {
		path = ""
	}

	getText, _ := loreley.CompileAndExecuteToString(
		`{bold}{fg 15}{bg 40} GET  {from "" 0}{reset}`,
		nil,
		nil,
	)
	postText, _ := loreley.CompileAndExecuteToString(
		`{bold}{fg 15}{bg 21} POST {from "" 0}{reset}`,
		nil,
		nil,
	)
	headText, _ := loreley.CompileAndExecuteToString(
		`{bold}{fg 15}{bg 53} HEAD {from "" 0}{reset}`,
		nil,
		nil,
	)
	putText, _ := loreley.CompileAndExecuteToString(
		`{bold}{fg 15}{bg 208} PUT  {from "" 0}{reset}`,
		nil,
		nil,
	)
	delText, _ := loreley.CompileAndExecuteToString(
		`{bold}{fg 15}{bg 160} DEL  {from "" 0}{reset}`,
		nil,
		nil,
	)
	
	if(status == "200" || status == "201") {
		status = color.GreenString(status)
	} else if(status == "404" || status == "500" || status == "403") {
		status = color.RedString(status)
	}

	//finally print the logs
	if(method == "GET") {
		mapLock.RLock()
		l.printf("%s %s - %s | %v | %4v | %s | %s \n", getText, timed, date, status, latency, ip, path)
		mapLock.RUnlock()
	} else if method == "POST" {
		mapLock.RLock()
		l.printf("%s %s - %s | %v | %4v | %s | %s \n", postText, timed, date, status, latency, ip, path)
		mapLock.RUnlock()
	} else if method == "PUT" {
		mapLock.RLock()
		l.printf("%s %s - %s | %v | %4v | %s | %s \n", putText, timed, date, status, latency, ip, path)
		mapLock.RUnlock()
	} else if method == "HEAD" {
		mapLock.RLock()
		l.printf("%s %s - %s | %v | %4v | %s | %s \n", headText, timed, date, status, latency, ip, path)
		mapLock.RUnlock()
	} else if method == "DELETE" {
		mapLock.RLock()
		l.printf("%s %s - %s | %v | %4v | %s | %s \n", delText, timed, date, status, latency, ip, path)
		mapLock.RUnlock()
	} else {
		mapLock.RLock()
		l.printf("%s - %s %v %4v %s %s %s \n", timed, date, status, latency, ip, method, path)
		mapLock.RUnlock()
	}

}

func rightPad2Len(s string, padStr string, overallLen int) string{
	var padCountInt int
	padCountInt = 1 + ((overallLen-len(padStr))/len(padStr))
	var retStr =  s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

func leftPad2Len(s string, padStr string, overallLen int) string{
	var padCountInt int
	padCountInt = 1 + ((overallLen-len(padStr))/len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr)-overallLen):]
}

func (l *loggerMiddleware) printf(format string, a ...interface{}) {
	if l.config.EnableColors {
		l.Logger.Otherf(format, a...)
	} else {
		l.Logger.Printf(format, a...)
	}
}

// New returns the logger middleware
// receives two parameters, both of them optionals
// first is the logger, which normally you set to the 'iris.Logger'
// if logger is nil then the middlewares makes one with the default configs.
// second is optional configs(logger.Config)
func New(theLogger *logger.Logger, cfg ...Config) iris.HandlerFunc {
	if theLogger == nil {
		theLogger = logger.New(logger.DefaultConfig())
	}
	c := DefaultConfig().Merge(cfg)
	l := &loggerMiddleware{Logger: theLogger, config: c}

	return l.Serve
}
