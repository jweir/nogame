package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"
)

// opens or creates an empty file for host configuration
func hostsFile() string {
	u, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	dir := fmt.Sprintf("%s/Applications/nogame/", u.HomeDir)
	err = os.MkdirAll(dir, 0777)

	if err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("%shosts.txt", dir)
	file, err := os.Open(filename)

	if err != nil {
		file, err = os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
	}

	return file.Name()
}

func Hosts() []string {
	str, e := ioutil.ReadFile(hostsFile())

	if e != nil {
		log.Fatal(e)
	}

	hosts := strings.Split(strings.TrimSpace(string(str)), "\n")
	return hosts
}

func BlockedHosts() *regexp.Regexp {
	return regexp.MustCompile(strings.Join(Hosts(), "|"))
}
