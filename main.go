package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var camera *Camera;

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

	// TODO: Cert & Key
	log.Fatal(http.ListenAndServe(":80", nil))
}
