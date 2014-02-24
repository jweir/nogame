package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const AllowedTime string = "1m"
const UnlockIn string = "18h"

type BlockClock struct {
	LockAt   time.Time
	UnlockAt time.Time
}

func (bl *BlockClock) Set(t time.Time) *BlockClock {
	half, _ := time.ParseDuration(AllowedTime)
	day, _ := time.ParseDuration(UnlockIn)

	bl.LockAt = t.Add(half)
	bl.UnlockAt = t.Add(day)
	return bl
}

func (bl *BlockClock) Update() {
	if bl.UnlockAt.Before(time.Now()) {
		bl.Set(time.Now())
		log.Printf("start block at %v\n", bl.LockAt)
	}
}

func (bl *BlockClock) Allow() bool {
	return bl.LockAt.After(time.Now())
}

func (bl *BlockClock) Block(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	log.Printf("blocking %s", r.URL.Host)
	bl.Update()
	if !bl.Allow() {
		r.URL.Host = "khanacademy.org"
		r.URL.Path = "/"
	}
	return r, nil
}

// configuration
func conf(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if r.URL.Path == "/exit" {
		os.Exit(1)
		return r, nil
	} else {
		return r, goproxy.NewResponse(r,
			goproxy.ContentTypeText,
			http.StatusOK,
			fmt.Sprintf(" %s", Hosts()))
	}
}

func Hosts() []string {
	str, e := ioutil.ReadFile("hosts")

	if e != nil {
		str = []byte("r2games.com")
		// log.Fatalf("%s", e)
	}

	return strings.Split(strings.TrimSpace(string(str)), "\n")
}

func BlockedHosts() *regexp.Regexp {
	return regexp.MustCompile(strings.Join(Hosts(), "|"))
}

func main() {
	bl := BlockClock{}
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest(goproxy.ReqHostMatches(BlockedHosts())).DoFunc(bl.Block)
	proxy.OnRequest(goproxy.DstHostIs("localhost")).DoFunc(conf)
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
