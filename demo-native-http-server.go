package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"crypto/rand"
	"syscall"
	"os/signal"

)
// to log all web requests.
var weblogger *log.Logger

// generateID uses rand from crypto module to generate random value
// into hexadecimal mode this value will be used as request id.
func generateID() string {

	// randomly fill the 8 capacity slice of bytes
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// use current number of nanoseconds since January 1, 1970 UTC
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// simple handler for /welcome URI.
func welcome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "simple plain welcome.")
}

// middleware that build request id and log useful details.
func logRequestMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get unique id for this request.
		requestId := generateID()
		log.Printf("received new request  [id: %s] - logging ...", requestId)
		weblogger.Printf("[id: %s] [ip: %s] [method: %s] [url: %s] [browser: %s]", requestId, r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
		next.ServeHTTP(w, r)
		log.Printf("completed new request [id: %s] - thanks ...", requestId)
	})
}

// starts the web server.
func startWebServer(exit <- chan struct{}) {
	const address = "127.0.0.1:8080"
	
	router := http.NewServeMux()
	router.Handle("/", http.NotFoundHandler())
	router.HandleFunc("/welcome", welcome)

	webserver := &http.Server{
		Addr:         address,
		Handler:      logRequestMiddleware(router),
		ErrorLog:     weblogger,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// goroutine in charge of shutting down the server when triggered.
	go func() {
		// wait until close or something comes in.
		<-exit
		log.Printf("shutting down the web server ... please wait for 45 secs max")
		ctx, _ := context.WithTimeout(context.Background(), 45*time.Second)

		// Shutdown gracefully shuts down the server without interrupting any active connections.
		// Shutdown works by first closing all open listeners, then closing all idle connections,
		// and then waiting indefinitely for connections to return to idle and then shut down.
		// If the provided context expires before the shutdown is complete, Shutdown returns the
		// context's error, otherwise it returns any error returned from closing the Server's
		// underlying Listener(s). *** from official documentation.
		// we can use webserver.Close() to immediately close it.
		if err := webserver.Shutdown(ctx); err != nil {
			// error due to closing listeners, or context timeout.
			log.Printf("failed to shutdown gracefully the web server - errmsg: %v", err)
			if err == context.DeadlineExceeded {
				log.Printf("the web server did not shutdown before 45 secs deadline.")
			} else {
				log.Printf("an error occured when closing underlying listeners.")
			}

			return
		}

		// err = nil - successfully shutdown the server.
		log.Printf("the web server was successfully shutdown down.")

	}()

	log.Println("web server is starting ...")
	if err := webserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start the web server on %s - errmsg: %v\n", address, err)
	}
}

// handleSignal is a function that process SIGTERM from kill command or CTRL-C or more.
func handleSignal(exit chan struct{}) {

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
		syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, os.Kill)

	// block until something comes in.
	signalType := <-sigch
	signal.Stop(sigch)
	fmt.Println("received signal type: ", signalType)
	fmt.Println("exit command received. exiting ...")
	// below triggers exit
	close(exit)
	return
}

func main() {
	weblogger = log.New(os.Stdout, "[request] ", log.LstdFlags)
	exit := make(chan struct{}, 1)
	go handleSignal(exit)
	startWebServer(exit)

}

//====== Outputs on server.

// [ 2:11:23] {nxos-geek}:~$ go run http-web.go
// 2021/08/26 02:15:45 web server is starting ...
// 2021/08/26 02:15:48 received new request  [id: 9c4565b075528b7f] - logging ...
// [request] 2021/08/26 02:15:48 [id: 9c4565b075528b7f] [ip: 127.0.0.1:63448] [method: GET] [url: /] [browser: curl/7.55.1]
// 2021/08/26 02:15:48 completed new request [id: 9c4565b075528b7f] - thanks ...
// 2021/08/26 02:15:52 received new request  [id: b56b0ea8c72dfcd6] - logging ...
// [request] 2021/08/26 02:15:52 [id: b56b0ea8c72dfcd6] [ip: 127.0.0.1:63451] [method: GET] [url: /welcome] [browser: curl/7.55.1]
// 2021/08/26 02:15:52 completed new request [id: b56b0ea8c72dfcd6] - thanks ...
// 2021/08/26 02:15:57 received new request  [id: 7f9e1905baf425b6] - logging ...
// [request] 2021/08/26 02:15:57 [id: 7f9e1905baf425b6] [ip: 127.0.0.1:63461] [method: GET] [url: /] [browser: curl/7.55.1]
// 2021/08/26 02:15:57 completed new request [id: 7f9e1905baf425b6] - thanks ...
// 2021/08/26 02:16:01 received new request  [id: 17ba4ecf3741bfb0] - logging ...
// [request] 2021/08/26 02:16:01 [id: 17ba4ecf3741bfb0] [ip: 127.0.0.1:63467] [method: GET] [url: /welcome] [browser: curl/7.55.1]
// 2021/08/26 02:16:01 completed new request [id: 17ba4ecf3741bfb0] - thanks ...
// received signal type:  interrupt
// exit command received. exiting ...
// 2021/08/26 02:16:05 shutting down the web server ... please wait for 45 secs max

// [ 2:16:05] {nxos-geek}:~$

//====== Outputs on client.
// [ 2:11:18] {nxos-geek}:~$ curl localhost:8080/
// 404 page not found

// [ 2:15:48] {nxos-geek}:~$ curl localhost:8080/welcome
// simple plain welcome.
// [ 2:15:52] {nxos-geek}:~$ curl localhost:8080/
// 404 page not found

// [ 2:15:57] {nxos-geek}:~$ curl localhost:8080/welcome
// simple plain welcome.
// [ 2:16:01] {nxos-geek}:~$