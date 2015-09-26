package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
)

/*
#cgo pkg-config: GraphicsMagick

#include "resize.h"
#include <stdlib.h>
*/
import "C"

func InitMagick() {
	C.InitializeMagick(nil)

}

func CloseMagick() {
	C.DestroyMagick()
}

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
	Dimensions

	collection, name string
	force            bool
}

func ParseResizeRequest(vars map[string]string, query map[string][]string) Request {
	r := Request{}
	var err error
	r.collection = vars["collection"]
	r.width, err = strconv.Atoi(vars["width"])
	if err == nil {
		r.height, err = strconv.Atoi(vars["height"])
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

type ResizeHandler struct {
	Confs
}

func (h ResizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we panic, write a 400 and return. this sucks, but uh, its fine for now.
	defer Write400(w)

	request := ParseResizeRequest(mux.Vars(r), r.URL.Query())
	file_struct := NewDiskImage(h.filePrefix, request.collection, fmt.Sprintf("resize/%dx%d", request.width, request.height), request.name)
	fetched_file, resize := file_struct.read()

	var resized_bytes []byte
	var length int

	if resize || request.force {
		log.Print("resizing")
		length = len(fetched_file)
		blob := C.resize_image(unsafe.Pointer(&fetched_file[0]), (*C.size_t)(unsafe.Pointer(&length)), (C.int)(request.width), (C.int)(request.height), 0, 13, 1.0)
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

type CapRequest struct {
	dimension int

	cap_dimension    C.cap_dimension
	collection, name string
	force            bool
}

func ParseCapRequest(dimension C.cap_dimension, vars map[string]string, query map[string][]string) CapRequest {
	r := CapRequest{}
	var err error

	r.collection = vars["collection"]
	r.cap_dimension = dimension
	r.dimension, err = strconv.Atoi(vars["dimension"])
	if err != nil {
		panic("Couldn't parse dimension")
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

func capImage(w http.ResponseWriter, request CapRequest, filePrefix string) {
	file_struct := NewDiskImage(filePrefix, request.collection, fmt.Sprintf("cap/%d", request.dimension), request.name)
	fetched_file, resize := file_struct.read()

	var resized_bytes []byte
	var length int
	var err C.cap_image_error

	if resize || request.force {
		log.Print("resizing")
		length = len(fetched_file)
		blob := C.cap_image(
			unsafe.Pointer(&fetched_file[0]),
			(*C.size_t)(unsafe.Pointer(&length)),
			(*C.cap_image_error)(unsafe.Pointer(&err)),
			(C.int)(request.dimension),
			request.cap_dimension,
			0, 13, 1.0, 100000, 100000)
		defer C.free(blob)
		// TODO need to do some checking on err

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

type CapHandler struct {
	Confs
	dimension C.cap_dimension
}

func (h CapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer Write400(w)

	request := ParseCapRequest(h.dimension, mux.Vars(r), r.URL.Query())
	capImage(w, request, h.filePrefix)
}
