package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
)

/*
#cgo CFLAGS: -I/usr/local/include/GraphicsMagick
#cgo LDFLAGS: -L/usr/local/lib -l GraphicsMagick

#include "resize.h"
#include <stdlib.h>
*/
import "C"

type Dimensions struct {
	width, height int
}

func Write400(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Print(r)
		w.WriteHeader(http.StatusBadRequest)
	}
}

type Request struct {
	collection, name string
	dim              Dimensions
	force            bool
}

func ParseRequest(vars map[string]string, query map[string][]string) Request {
	r := Request{}
	var err error
	r.collection = vars["collection"]
	r.dim.width, err = strconv.Atoi(vars["width"])
	if err == nil {
		r.dim.height, err = strconv.Atoi(vars["height"])
		if err != nil {
			log.Print("Couldn't parse height")
		}
	} else {
		log.Print("Couldn't parse wid")
	}
	r.name = vars["name"]
	if arr := query["force"]; len(arr) > 0 {
		r.force, err = strconv.ParseBool(arr[0])
		if err != nil {
			r.force = false
		}
	}

	return r
}

func resizeHandler(w http.ResponseWriter, r *http.Request) {
	defer Write400(w)

	request := ParseRequest(mux.Vars(r), r.URL.Query())
	file_struct := NewDiskImage(Conf.file_prefix, request.collection, fmt.Sprintf("%dx%d", request.dim.width, request.dim.height), request.name)
	dimensions := request.dim
	fetched_file, resize := file_struct.read()

	var resized_bytes []byte
	var length int

	if resize || request.force {
		log.Print("resizing")
		length = len(fetched_file)
		blob := C.resize_image(unsafe.Pointer(&fetched_file[0]), (*C.size_t)(unsafe.Pointer(&length)), (C.int)(dimensions.width), (C.int)(dimensions.height), 0, 13, 1.0)
		defer C.free(blob)

		// copy to go; I can make this faster with some "internal" things, but that can come later
		resized_bytes = C.GoBytes(blob, (C.int)(length))

		go file_struct.write(resized_bytes)
	} else {
		resized_bytes = fetched_file
		length = len(fetched_file)
	}

	// TODO(jake) return the actual type
	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("Content-Length", strconv.Itoa(length))
	w.Header().Add("Last-Modified", time.Now().Format(time.RFC1123))
	w.WriteHeader(http.StatusOK)
	w.Write(resized_bytes)
}

func capHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("capHandler")
}

var Conf Confs = Confs{"images/", 8000}

type Confs struct {
	file_prefix string
	port        int
}

func main() {
	C.CreateMagick()

	/* here we ensure that go's signal handlers don't interfere. We have to shut
	down graphicsmagick correctly or crash */
	signal_chan := make(chan os.Signal, 1)
	// Blow away go's handlers
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(signal_chan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signal_chan
		// clean up graphicsmagick's memory / event loops
		C.DestroyMagick()
		os.Exit(1)
	}()

	r := mux.NewRouter()

	// /img/collection/dimensions/name?flags
	r.HandleFunc("/img/{collection}/cap{dimensions}/{name}", capHandler)
	// todo these
	//r.HandleFunc("/img/{collection}/width{dimensions}/{name}", widthHandler)
	//r.HandleFunc("/img/{collection}/height{dimensions}/{name}", heightHandler)
	r.HandleFunc("/img/{collection}/{width}x{height}/{name}", resizeHandler)
	log.Print("Starting imageservice")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Conf.port), r))
}
