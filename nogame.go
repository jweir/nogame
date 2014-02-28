package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"regexp"
	"time"
)

const AllowedTime time.Duration = time.Minute * 30
const Timeout time.Duration = time.Hour * 18

type BlockClock struct {
	// When to start blocking the hosts
	LockAt time.Time

	// When to unlock the hosts
	UnlockAt time.Time

	// What hosts are blocked
	Hosts []string

	// regex to unlock the hosts
	HostsPattern *regexp.Regexp
}

func (bl *BlockClock) Set(t time.Time) *BlockClock {
	bl.LockAt = t.Add(AllowedTime)
	bl.UnlockAt = t.Add(Timeout)
	return bl
}

func (bl *BlockClock) startTimer() bool {
	if bl.UnlockAt.Before(time.Now()) {
		bl.Set(time.Now())
		return true
	}

	return false
}

func (bl *BlockClock) locked() bool {
	return bl.LockAt.After(time.Now())
}

func (bl *BlockClock) Block(r *http.Request) *http.Request {
	if bl.startTimer() {
		log.Printf("will block %s at %s", r.URL.Host, bl.LockAt)
	}

	if !bl.locked() {
		r.URL.Host = "khanacademy.org"
		r.URL.Path = "/"
	}
	return r
}

func (bl *BlockClock) blockedHost(host string) bool {
	return bl.HostsPattern.Match([]byte(host))
}

func (bl *BlockClock) checkHost(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if bl.blockedHost(r.URL.Host) {
		time.Sleep(time.Second * 4)
		return bl.Block(r), nil
	} else {
		return r, nil
	}
}

// display the configuration
func (bl *BlockClock) conf(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	bl.update()

	return r, goproxy.NewResponse(r,
		goproxy.ContentTypeText,
		http.StatusOK,
		fmt.Sprintf("hosts %s\nstart blocking at: %s", bl.Hosts, bl.LockAt))
}

func Create() *BlockClock {
	bl := &BlockClock{}
	bl.update()

	return bl
}

func (bl *BlockClock) update() *BlockClock {
	bl.Hosts = Hosts()
	bl.HostsPattern = BlockedHosts()

	log.Printf("blocking %v\n", bl.Hosts)
	return bl
}

func main() {
	bl := Create()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest(goproxy.DstHostIs("nogame")).DoFunc(bl.conf)
	proxy.OnRequest().DoFunc(bl.checkHost)
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
