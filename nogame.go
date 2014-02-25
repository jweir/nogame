package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"regexp"
	"time"
)

const AllowedTime string = "30m"
const UnlockIn string = "18h"

type BlockClock struct {
	LockAt       time.Time
	UnlockAt     time.Time
	Port         int
	Path         string
	AllowedTime  time.Duration
	Timeout      time.Duration
	Hosts        []string
	HostsPattern *regexp.Regexp
}

func (bl *BlockClock) Set(t time.Time) *BlockClock {
	bl.LockAt = t.Add(bl.AllowedTime)
	bl.UnlockAt = t.Add(bl.Timeout)
	return bl
}

func (bl *BlockClock) CheckLock() bool {
	if bl.UnlockAt.Before(time.Now()) {
		bl.Set(time.Now())
		log.Printf("start block at %v\n", bl.LockAt)
		return true
	}

	return false
}

func (bl *BlockClock) Allow() bool {
	return bl.LockAt.After(time.Now())
}

func (bl *BlockClock) Block(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if bl.CheckLock() {
		log.Printf("blocking %s", r.URL.Host)
	}
	if !bl.Allow() {
		r.URL.Host = "khanacademy.org"
		r.URL.Path = "/"
	}
	return r, nil
}

func (bl *BlockClock) CheckHostForBlocking(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
  if bl.HostsPattern.Match([]byte(r.URL.Host)){
    return bl.Block(r, ctx);
  } else {
    return r, nil
  }
}

// display the configuration
func (bl *BlockClock) conf(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	log.Println("conf...")
	if r.URL.Path == "/update" {
		bl.Update()
	}

	return r, goproxy.NewResponse(r,
		goproxy.ContentTypeText,
		http.StatusOK,
		fmt.Sprintf("hosts %s\nstart blocking at: %s", bl.Hosts, bl.LockAt))
}


func Create() *BlockClock {
	half, _ := time.ParseDuration("30m")
	day, _ := time.ParseDuration("18h")

	bl := &BlockClock{
		AllowedTime: half,
		Timeout:     day,
	}

	bl.Update()

	return bl
}

func (bl *BlockClock) Update() *BlockClock {
	bl.Hosts = Hosts()
	bl.HostsPattern = BlockedHosts()

  log.Printf("blocking %v\n",bl.Hosts)
	return bl
}

func main() {
	bl := Create()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest(goproxy.DstHostIs("localhost")).DoFunc(bl.conf)
	proxy.OnRequest().DoFunc(bl.CheckHostForBlocking)
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
