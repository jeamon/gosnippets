package main

// This is a small go-based nice demonstration of loading multiple ip addresses from pipe input data and
// from any number of files passed as program arguments. It processes all data and store on valid ips.

// Version  : 1.0
// Author   : Jerome AMON
// Created  : 19 November 2021

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// isValidIP takes a supposed ip address and returns true if it is valid.
func isValidIP(ip string) bool {
	if net.ParseIP(ip) != nil {
		return true
	}

	return false
}

// loadInfos loads data piped and from all files passed
// as program arguments and fill the pointer of slice
// with valid IP addresses.
func loadInfos(ips *[]string) {

	var entries []string

	// retrieve standard input info.
	fi, _ := os.Stdin.Stat()

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// there is data is from pipe, so grab the
		// full content and build a list of entries.
		content, _ := ioutil.ReadAll(os.Stdin)
		entries = strings.Split(string(content), "\n")

	}

	// parse any files content.
	filenames := os.Args[1:]

	if len(filenames) > 0 {
		// for each valid file path, grab its full
		// content and build a list of entries.
		var lines []string
		for _, file := range filenames {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			// construct the list based on "\n" as sep.
			// then add lines content to entries list.
			lines = strings.Split(string(content), "\n")
			entries = append(entries, lines...)
		}
	}

	if len(entries) == 0 {
		// no data input.
		return
	}

	// keep only valid IP addresses.
	for _, e := range entries {
		if isValidIP(strings.TrimSpace(e)) {
			*ips = append(*ips, strings.TrimSpace(e))
		}
	}
}

func main() {
	// hold all ips to process.
	var ips []string
	loadInfos(&ips)
	fmt.Println(ips)
}

/*

// create a file name ips.txt with below content.

127.0.0.1
127.0.0.1
8.8.8.8
8.8.8.8

// Then run the program as below.

~$ type ips.txt | go run load-ip-addresses.go ips.txt ips.txt
[127.0.0.1 127.0.0.1 8.8.8.8 8.8.8.8 127.0.0.1 127.0.0.1 8.8.8.8 8.8.8.8 127.0.0.1 127.0.0.1 8.8.8.8 8.8.8.8]

*/
