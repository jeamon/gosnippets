package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// this is a simple go program that demonstrates how to spin up HTTP & HTTPS servers at the same time
// and force all HTTP traffic to move to HTTPS - all done with customized TLS configurations.
// you will find at the end of this file openssl commands to generate your self-signed certificates.
// this program expects to load key and self-signed certificate from child directory named "certificates".

// http server address.
const httpaddress = "127.0.0.1:8080"

// https server address.
const httpsaddress = "127.0.0.1:8443"

// redirectToHTTPS forces to use HTTPS by redirecting the request.
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Printf("redirecting request from http to https with hsts setup.")
	// Strict Transport Security (aka HSTS) is so (if supported by the browser) all present and
	// future subdomains communications will be sent over HTTPS. Test with Max-age time of 1 hour.
	w.Header().Set("Strict-Transport-Security", "max-age=3600; includeSubDomains")
	// wants to close this current connection and self-initiate a redirect (code 301) to HTTPS.
	w.Header().Set("Connection", "close")
	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
}

// simple handler for /Hello URI.
func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "simple hello user.")
}

// handleSignal is a function that process SIGTERM from kill command or CTRL-C or more.
func handleSignal(done chan struct{}, httpwebserver *http.Server, httpswebserver *http.Server) {

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
		syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, os.Kill)

	// block until something comes in.
	signalType := <-sigch
	signal.Stop(sigch)
	fmt.Printf("received signal type [%v] exiting ...", signalType)

	// immediately closes all active net.Listeners and any connections in
	// state StateNew, StateActive, or StateIdle.
	log.Println("shutting down the http web server - please wait ...")
	if err := httpwebserver.Close(); err != nil {
		log.Printf("an error occured when closing http server - errmsg : %v", err)
	}

	log.Println("shutting down the https web server - please wait ...")
	if err := httpswebserver.Shutdown(context.Background()); err != nil {
		log.Println("an error occured when shutting down https server.")
	}
	log.Println("completed shutting down http & https web servers")
	close(done)
	return
}

// setupHTTPSServer constructs the custom HTTPS server based on TLS details.
func setupHTTPSServer() *http.Server {
	// load server certificate and key.
	serverCerts, err := tls.LoadX509KeyPair("certificates/server.rsa.crt", "certificates/server.rsa.key")
	if err != nil {
		log.Println("failed to load server certificate and key files - errmsg : %v", err)
		os.Exit(1)
	}
	// TLS configuration details - https://pkg.go.dev/crypto/tls#Config
	tlsConfig := &tls.Config{
		// handshake with minimum TLS 1.2. MaxVersion is 1.3.
		MinVersion: tls.VersionTLS12,
		// server certificates (key with self-signed certs).
		Certificates: []tls.Certificate{serverCerts},
		// https://pkg.go.dev/crypto/tls#ClientAuthType.
		ClientAuth: tls.RequestClientCert,
		// CAs to be used to authenticate clients certificates.
		// ClientCAs: ClientCAs,
		// elliptic curves that will be used in an ECDHE handshake, in preference order.
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.X25519, tls.CurveP256},

		// CipherSuites is a list of enabled TLS 1.0–1.2 cipher suites. The order of
		// the list is ignored. Note that TLS 1.3 ciphersuites are not configurable.
		CipherSuites: []uint16{
			// ECDHE is the algorithm for Key exchange.
			// RSA or ECDSA for keys authentication.
			// AES256 or AES128 GCM/CBC for encryption.
			// SHA384 or SHA256 or SHA for hashing.

			// TLS v1.2 - ECDSA-based keys cipher suites.
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,

			// TLS v1.2 - RSA-based keys cipher suites.
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// TLS v1.3 - some strong cipher suites.
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},

		// Deprected: PreferServerCipherSuites is ignored.
		PreferServerCipherSuites: true,
	}

	// base http router.
	router := http.NewServeMux()

	// simpple inline handler function for root URI.
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello From TLS Server.\n")
	})
	// another URI with its handler.
	router.HandleFunc("/hello", hello)

	// non-secure and secure webservers parameters.
	httpsserver := &http.Server{
		Addr:         httpsaddress,
		Handler:      router,
		ErrorLog:     log.New(os.Stdout, "[https] ", log.LstdFlags),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig:    tlsConfig,
	}

	return httpsserver
}

func main() {
	// have an instance of our custom HTTPS server.
	httpsserver := setupHTTPSServer()
	// basic HTTP server with redirect to HTTPS.
	httpserver := &http.Server{
		Addr:         httpaddress,
		Handler:      http.HandlerFunc(redirectToHTTPS),
		ErrorLog:     log.New(os.Stdout, "[http ] ", log.LstdFlags),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// goroutine to handle exit signals and initiate servers closure.
	done := make(chan struct{}, 1)
	go handleSignal(done, httpserver, httpsserver)

	log.Printf("starting non-secure web server (http) at %s ...", httpsaddress)
	// http server which redirects every http requests to https server.
	go func() {
		if err := httpserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// ListenAndServeTLS always returns a non-nil error. Shutdown or Close triggers ErrServerClosed.
			log.Printf("failed to start http web server on %s - errmsg: %v\n", httpaddress, err)
		}
	}()

	log.Printf("starting secure web server (https) at %s ...", httpsaddress)
	// start the HTTPS web server and block until error event happen.
	if err := httpsserver.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		// ListenAndServeTLS always returns a non-nil error. Shutdown or Close triggers ErrServerClosed.
		log.Printf("failed to start https web server on %s - errmsg: %v\n", httpsaddress, err)
	}

	// block to make sure both servers are turned off.
	<-done
}

// USEFUL DETAILS TO GENERATE CERTIFICATES with OpenSSL 1.1.1c

//== RSA-Based keys - Self Signed Certificate Generation (key - recommendation key ≥ 2048)
// openssl req -x509 -nodes -newkey rsa:4096 -keyout server.rsa.key -out server.rsa.crt -days 365

//== ECDSA-Based keys - Self Signed Certificate Generation (key - recommendation key ≥ secp384r1)
// To list ECDSA supported curves use following command : openssl ecparam -list_curves
// openssl req -new -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -x509 -nodes -days 365 -out server.ecdsa.crt -keyout server.ecdsa.key
// openssl req -x509 -nodes -newkey ec:secp384r1 -keyout server.ecdsa.key -out server.ecdsa.crt -days 365
// openssl req -x509 -nodes -newkey ec:<(openssl ecparam -name secp384r1) -keyout server.ecdsa.key -out server.ecdsa.crt -days 365

//== If needed to generate the Certficate Signing Request (CSR)
// openssl req -new -sha256 -key server.key -out server.csr
// openssl x509 -req -sha256 -in server.csr -signkey server.key -out server.crt -days 3650

// TLS   : Transport Layer Security
// ECDHE : Ephemeral Elliptic Curve Diffie-Hellman
// RSA   : Rivest–Shamir–Adleman
// ECDSA : Elliptic Curve Digital Signature Algorithm
// GCM   : Galois Counter Mode
// SHA   : Secure Hash Algorithm

// Function 					Algorithms
//-----------------------------------------------------------------
// Key Exchange 				RSA, Diffie-Hellman, ECDH, SRP, PSK
//-----------------------------------------------------------------
// Authentication 				RSA, DSA, ECDSA
//-----------------------------------------------------------------
// Bulk Ciphers (Encryption)	RC4, 3DES, AES
//-----------------------------------------------------------------
// Message Auth (Hashing) 		HMAC-SHA256, HMAC-SHA1, HMAC-MD5
//-----------------------------------------------------------------
