package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/theevocater/adminz"
)

/*
#cgo pkg-config: GraphicsMagick

#include "resize.h"
*/
import "C"

var Conf Confs

type Confs struct {
	filePrefix          string
	port                string
	maxWidth, maxHeight int
}

func readSettings() []string {
	c := Confs{}

	flag.StringVar(&c.filePrefix, "filePrefix", "", "full path to the file sandbox you wish to serve")

	flag.StringVar(&c.port, "port", "8000", "listen port")

	flag.IntVar(&c.maxHeight, "maxHeight", 10000, "maximum height of image in pixels")
	flag.IntVar(&c.maxWidth, "maxWidth", 10000, "maximum width of image in pixels")

	flag.Parse()
	Conf = c

	return flag.Args()
}

func main() {
	InitMagick()

	/* here we ensure that go's signal handlers don't interfere. We have to shut
	down graphicsmagick correctly or crash */
	signal_chan := make(chan os.Signal, 1)
	// Blow away go's handlers
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(signal_chan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signal_chan
		// clean up graphicsmagick's memory / event loops
		CloseMagick()
		adminz.Stop()
		os.Exit(1)
	}()

	r := mux.NewRouter()

	readSettings()

	capHandler := CapHandler{Conf, C.CAP}
	r.Handle("/img/{collection}/cap{dimension}/{name}", capHandler)
	r.Handle("/img/{collection}/cap/{dimension}/{name}", capHandler)

	widthHandler := CapHandler{Conf, C.WIDTH}
	r.Handle("/img/{collection}/width{dimension}/{name}", widthHandler)
	r.Handle("/img/{collection}/width/{dimension}/{name}", widthHandler)

	heightHandler := CapHandler{Conf, C.HEIGHT}
	r.Handle("/img/{collection}/height{dimension}/{name}", heightHandler)
	r.Handle("/img/{collection}/height/{dimension}/{name}", heightHandler)

	r.Handle("/img/{collection}/{width}x{height}/{name}", ResizeHandler{Conf})
	http.Handle("/", r)

	log.Print("Starting imageservice")

	adminz.Init(Conf.port, func() string { return "{ \"a\": 1 }" })

	log.Printf("Listening on port %s", Conf.port)
	log.Printf("Reading images from disk at %s", Conf.filePrefix)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", Conf.port), nil))
}
