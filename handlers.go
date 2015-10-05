package main

import (
	"log"
	"mime"
	"net/http"
	"path"
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

type Dimensions struct {
	width, height int
}

func Write400(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Print(r)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func WriteImage(name string, output []byte, length int, w http.ResponseWriter) {
	w.Header().Add("Content-Type", mime.TypeByExtension(path.Ext(name)))
	w.Header().Add("Content-Length", strconv.Itoa(length))
	w.Header().Add("Last-Modified", time.Now().Format(time.RFC1123))
	w.WriteHeader(http.StatusOK)
	w.Write(output)
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
	imageCollections map[string]ImageCollection
	*ImageFactory
}

func (h ResizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we panic, write a 400 and return. this sucks, but uh, its fine for now.
	defer Write400(w)

	request := ParseResizeRequest(mux.Vars(r), r.URL.Query())
	collection := h.imageCollections[request.collection]
	resized := collection.GetResized(request.name, request.width, request.height)
	original := collection.GetOriginal(request.name)
	file := h.NewImage(resized, original)
	input, resize := file.Read()

	var output []byte
	var length int

	if resize || request.force {
		log.Print("resizing")
		length = len(input)
		blob := C.resize_image(unsafe.Pointer(&input[0]), (*C.size_t)(unsafe.Pointer(&length)), (C.int)(request.width), (C.int)(request.height), 0, 13, 1.0)
		defer C.free(blob)

		// copy to go; I can make this faster with some "internal" things, but that can come later
		output = C.GoBytes(blob, (C.int)(length))

		go file.Write(output)
	} else {
		output = input
		length = len(input)
	}

	WriteImage(request.name, output, length, w)
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

type CapHandler struct {
	Confs
	imageCollections map[string]ImageCollection
	*ImageFactory
	dimension C.cap_dimension
}

func (h CapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer Write400(w)

	request := ParseCapRequest(h.dimension, mux.Vars(r), r.URL.Query())
	collection := h.imageCollections[request.collection]
	var resized string

	switch h.dimension {
	case C.CAP:
		resized = collection.GetCapped(request.name, request.dimension)
	case C.WIDTH:
		resized = collection.GetWidth(request.name, request.dimension)
	case C.HEIGHT:
		resized = collection.GetHeight(request.name, request.dimension)
	}
	original := collection.GetOriginal(request.name)
	file := h.NewImage(resized, original)
	capImage(w, request, file)
}

func capImage(w http.ResponseWriter, request CapRequest, file ImageFile) {
	input, resize := file.Read()

	var output []byte
	var length int
	var err C.cap_image_error

	if resize || request.force {
		log.Print("resizing")
		length = len(input)
		blob := C.cap_image(
			unsafe.Pointer(&input[0]),
			(*C.size_t)(unsafe.Pointer(&length)),
			(*C.cap_image_error)(unsafe.Pointer(&err)),
			(C.int)(request.dimension),
			request.cap_dimension,
			0, 13, 1.0, 100000, 100000)
		defer C.free(blob)

		if err != C.CAP_IMAGE_ERROR_OK {
			panic("Failed to resize")
		}

		// copy to go; I can make this faster with some "internal" things, but that can come later
		output = C.GoBytes(blob, (C.int)(length))

		go file.Write(output)
	} else {
		output = input
		length = len(input)
	}

	WriteImage(request.name, output, length, w)
}
