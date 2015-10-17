package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/foursquare/fsgo/adminz"
	"github.com/gorilla/mux"
)

/*
#cgo pkg-config: GraphicsMagick

#include "resize.h"
*/
import "C"

var Conf Confs

var adminzEndpoints *adminz.Adminz

var imageCollections map[string]ImageCollection

type Confs struct {
	Port                string
	MaxWidth, MaxHeight int

	// Settings needed for S3
	S3         bool
	BucketName string

	CollectionsPath string
}

func readSettings() []string {
	c := Confs{}

	flag.BoolVar(&c.S3, "s3", false, "Use s3 instead of the local filesystem")
	flag.StringVar(&c.BucketName, "bucketName", "", "Name of S3 bucket to read from")

	flag.StringVar(&c.CollectionsPath, "collections", "", "json file of image collections")

	flag.StringVar(&c.Port, "port", "8000", "listen port")

	flag.IntVar(&c.MaxHeight, "maxHeight", 10000, "maximum height of image in pixels")
	flag.IntVar(&c.MaxWidth, "maxWidth", 10000, "maximum width of image in pixels")

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
		os.Exit(1)
	}()

	r := mux.NewRouter()

	readSettings()
	var factory *ImageFactory
	if Conf.S3 {
		factory = NewS3ImageFactory(Conf.BucketName)
	} else {
		factory = NewDiskImageFactory()
	}

	imageCollections, err := ParseImageCollections(Conf.CollectionsPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Print("Found collections: ")
	for k, _ := range imageCollections {
		log.Print(k)
	}

	handler := Handler{
		Confs:            Conf,
		imageCollections: imageCollections,
		ImageFactory:     factory,
	}

	capHandler := CapHandler{
		Handler:   handler,
		dimension: C.CAP,
	}
	r.Handle("/img/{collection}/cap{dimension}/{name}", capHandler)
	r.Handle("/img/{collection}/cap{dimension}/blur{blur}/{name}", capHandler)
	r.Handle("/img/{collection}/cap/{dimension}/{name}", capHandler)
	r.Handle("/img/{collection}/cap/{dimension}/blur{blur}/{name}", capHandler)

	widthHandler := CapHandler{
		Handler:   handler,
		dimension: C.WIDTH,
	}
	r.Handle("/img/{collection}/width{dimension}/{name}", widthHandler)
	r.Handle("/img/{collection}/width{dimension}/blur{blur}/{name}", widthHandler)
	r.Handle("/img/{collection}/width/{dimension}/{name}", widthHandler)
	r.Handle("/img/{collection}/width/{dimension}/blur{blur}/{name}", widthHandler)

	heightHandler := CapHandler{
		Handler:   handler,
		dimension: C.HEIGHT,
	}
	r.Handle("/img/{collection}/height{dimension}/{name}", heightHandler)
	r.Handle("/img/{collection}/height{dimension}/blur{blur}/{name}", heightHandler)
	r.Handle("/img/{collection}/height/{dimension}/{name}", heightHandler)
	r.Handle("/img/{collection}/height/{dimension}/blur{blur}/{name}", heightHandler)

	resizeHandler := ResizeHandler{
		Handler: handler,
	}
	r.Handle("/img/{collection}/{width}x{height}/{name}", resizeHandler)
	r.Handle("/img/{collection}/{width}x{height}/blur{blur}/{name}", resizeHandler)

	originalHandler := OriginalHandler{
		Handler: handler,
	}
	r.Handle("/img/{collection}/original/{name}", originalHandler)

	http.Handle("/", r)

	log.Print("Starting imageservice")

	adminzEndpoints = adminz.New()
	adminzEndpoints.KillfilePaths(adminz.Killfiles(Conf.Port))
	adminzEndpoints.Build()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", Conf.Port), nil))
}
