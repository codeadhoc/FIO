package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	log.Println("api service starting...")
	mux := http.NewServeMux()
	mux.HandleFunc("/mutual", hello)
	srv := http.Server{
		Addr:    ":8000",
		Handler: mux,
		TLSConfig: getTLSConfig("ecomm.kubesure.io", "/home/prashant/src/forkjoin/certs/ca.crt",
			tls.ClientAuthType(tls.VerifyClientCertIfGiven)),
	}
	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Println("shutting down server...")
			srv.Shutdown(ctx)
			<-ctx.Done()
		}
	}()
	//os.Getwd()
	if err := srv.ListenAndServeTLS("/home/prashant/src/forkjoin/certs/server.crt", "/home/prashant/src/forkjoin/certs/server.key"); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %s", err)
	}
}

func getTLSConfig(host, caCertFile string, certOpt tls.ClientAuthType) *tls.Config {
	var caCert []byte
	var err error
	var caCertPool *x509.CertPool
	if certOpt > tls.RequestClientCert {
		caCert, err = ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Fatal("Error opening cert file", caCertFile, ", error ", err)
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	return &tls.Config{
		ServerName: host,
		ClientAuth: certOpt,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12, // TLS versions below 1.2 are considered insecure - see https://www.rfc-editor.org/rfc/rfc7525.txt for details
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	//to simulate a deplyed response
	time.Sleep(5 * time.Second)
	w.WriteHeader(200)
	data := (time.Now()).String()
	log.Println("health ok")
	w.Write([]byte(data))
}
