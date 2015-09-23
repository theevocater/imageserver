package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

/*
#cgo CFLAGS: -I/usr/local/include/GraphicsMagick
#cgo LDFLAGS: -L/usr/local/lib -l GraphicsMagick

#include "resize.h"
#include <stdlib.h>
*/
import "C"

// /img/user/256x256/10029684-G5FUMQAZ10SCHGWV.jpg
// /img/collection/dimensions/name?flags
var handler_regxp = regexp.MustCompile("/img/([^/]+)/([^/]+)/([^/]+)")

type Dimensions struct {
	width, height int
}

func Write400(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Print(r)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func ParseDimensions(str string) Dimensions {
	split := strings.Split(str, "x")
	if len(split) == 2 {
		width, err := strconv.Atoi(split[0])
		if err != nil {
			panic("Couldn't parse dimensions")
		}
		height, err := strconv.Atoi(split[1])
		return Dimensions{width, height}
	}
	panic("Couldn't parse dimensions")
}

func ParseURL(url string) (string, string, string) {
	matches := handler_regxp.FindStringSubmatch(url)
	if len(matches) != 4 {
		panic("Couldn't match url")
	}
	collection := matches[1]
	dimensions := matches[2]
	name := matches[3]
	return collection, dimensions, name
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer Write400(w)

	bucket, dimensions_str, name := ParseURL(r.URL.Path)
	file_struct := NewDiskImage(Conf.file_prefix, bucket, dimensions_str, name)
	dimensions := ParseDimensions(dimensions_str)
	fetched_file, resize := file_struct.read()

	var resized_bytes []byte
	var length int

	if resize {
		log.Print("resizing")
		length = len(fetched_file)
		blob := C.resize_image(unsafe.Pointer(&fetched_file[0]), (*C.size_t)(unsafe.Pointer(&length)), (C.int)(dimensions.width), (C.int)(dimensions.height), 0, 13, 1.0)
		defer C.free(blob)

		// copy to go
		resized_bytes = C.GoBytes(blob, (C.int)(length))

		// is this safe?? I don't want to make too many copies
		go file_struct.write(resized_bytes)
	} else {
		resized_bytes = fetched_file
		length = len(fetched_file)
	}

	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("Content-Length", strconv.Itoa(length))
	w.Header().Add("Last-Modified", time.Now().Format(time.RFC1123))
	w.WriteHeader(http.StatusOK)
	w.Write(resized_bytes)
}

var Conf Confs = Confs{"images/", 8000}

type Confs struct {
	file_prefix string
	port        int
}

func main() {
	C.CreateMagick()
	defer C.DestroyMagick()
	// something is going on here with graphics magick causing it to crash w/
	// term :(
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, os.Interrupt)
	signal.Notify(signal_chan, syscall.SIGTERM)
	signal.Notify(signal_chan, syscall.SIGINT)
	go func() {
		<-signal_chan
		C.DestroyMagick()
		os.Exit(1)
	}()
	http.HandleFunc("/img/", handler)
	log.Print("Starting imageservice")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Conf.port), nil))
}
