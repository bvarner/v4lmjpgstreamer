package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var camera *Camera;

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://" + r.Host + r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	var err error = nil

	camTrigC   := make(chan time.Time, 1)

	// TODO: Load settings for input device.
	camera, err = NewCamera("dev/video0", camTrigC)
	if err != nil {
		fmt.Println("Camera not initialized: ", err);
	} else {
		defer camera.Close()
	}

	// Start the thread that triggers camera frame grabbing
	cameraPoller := time.NewTicker(40000 * time.Microsecond) // 25fps
	go func() {
		for t:= range cameraPoller.C {
			if camera.Initialized {
				camTrigC <- t
			}
		}
	}()

	fmt.Println("Setting up HTTP Server")

	// TODO: REST API For video device control
	http.HandleFunc("/", camera.ServeHTTP)

	cert := flag.String("cert", "/etc/ssl/certs/v4lmjpgstreamer.pem", "The certificate for this server.")
	certkey := flag.String("key", "/etc/ssl/certs/v4lmjpgstreamer-key.pem", "The key for the server cert.")

	_, certerr := os.Stat(*cert)
	_, keyerr := os.Stat(*certkey)

	if certerr == nil && keyerr == nil {
		fmt.Println("SSL Configuration set up.")
		go func() {
			log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(redirectTLS)));
		} ()
		log.Fatal(http.ListenAndServeTLS(":443", *cert, *certkey, nil))
	} else {
		fmt.Println("SSL Configuration not found.")
		log.Fatal(http.ListenAndServe(":80", nil))
	}
}
