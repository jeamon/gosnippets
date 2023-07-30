package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var dynamicRoutes map[string]string
var addRoutesMutex *sync.RWMutex
var latestStat os.FileInfo

// intialization
func init() {

	if dynamicRoutes == nil {
		dynamicRoutes = make(map[string]string)
	}

	addRoutesMutex = &sync.RWMutex{}
	loadDynamicRoutes() // initial loading of routes from file
}

// load and convert the json content into map
func loadDynamicRoutes() {

	routesFile, err := os.Open("dynamic-routes.json")
	if err != nil {
		log.Println("[ Eror ] Failed to load dynamic routes file. ErrMsg -", err)
		os.Exit(1)
	}
	defer routesFile.Close()
	latestStat, err = routesFile.Stat()
	if err != nil {
		log.Println("[ Eror ] Failed to get latest statistics of dynamic routes file. ErrMsg -", err)
		return
	}

	// convert the content into byte array
	routesBytes, err := ioutil.ReadAll(routesFile)
	if err != nil {
		log.Println("[ Eror ] Failed to convert dynamic routes into map. ErrMsg -", err)
		return
	}
	// lock the map to prevent race condition. enforce that
	// reading happening from another goroutines
	addRoutesMutex.Lock()

	// if needed to reflect same state as file
	// clean and recreate of the map like below
	dynamicRoutes = nil
	dynamicRoutes = make(map[string]string)

	// construct the map from the json content (byte array)
	err = json.Unmarshal([]byte(routesBytes), &dynamicRoutes)
	if err != nil {
		log.Println("[ Eror ] Failed to convert dynamic routes into map. ErrMsg -", err)
		return
	}
	addRoutesMutex.Unlock()

	// just displaying to check the content
	addRoutesMutex.RLock()
	for route, url := range dynamicRoutes {
		log.Println("route", route, "url:", url)
	}
	log.Printf("Current Number Of Routes Is : %d\n\n", len(dynamicRoutes))
	addRoutesMutex.RUnlock()
}

// check every interval minute and update if changes
func updateDynamicRoutes(interval int) {
	for {
		stat, err := os.Stat("dynamic-routes.json")
		if err != nil {
			log.Println("[ Eror ] Failed to get statistics of dynamic routes file. ErrMsg -", err)
		} else {
			if stat.Size() != latestStat.Size() || stat.ModTime() != latestStat.ModTime() {
				loadDynamicRoutes() // load only when size or latest modification time changed
			}
		}
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

// change this file dynamic-routes.json content and observe
func main() {
	doneChan := make(chan bool)
	go updateDynamicRoutes(2) // check to update each 2 min if any changes
	<-doneChan
}

/* function to call into your NOTFOUND HANDLER for checking dynamic routes
func checkDynamicRoutes(w http.ResponseWriter, r *http.Request) {

	addRoutesMutex.Rlock()
	if targetURL, found := dynamicRoutes[r.URL.String()]; found {
		addRoutesMutex.RUnlock()
		http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
		return
	} else {
		// not found routine goes here
		log.Println("unknown requested path - thank you.")
	}
}
*/
