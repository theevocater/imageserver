package main

import (
	"log"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/gorilla/mux"
)

/*
#cgo pkg-config: GraphicsMagick

#include "resize.h"
*/
import "C"

type capRequest struct {
	dimension int

	cap_dimension    C.cap_dimension
	collection, name string
	Modifiers
}

func parseCap(dimension C.cap_dimension, vars map[string]string, query map[string][]string) capRequest {
	r := capRequest{}
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

	r.Modifiers = parseModifiers(vars, query)

	return r
}

type CapHandler struct {
	Handler
	dimension C.cap_dimension
}

func (h CapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer Write400(w)

	request := parseCap(h.dimension, mux.Vars(r), r.URL.Query())
	collection := h.imageCollections[request.collection]
	var resized string

	switch h.dimension {
	case C.CAP:
		resized = collection.GetCapped(request.name, request.dimension, request.Modifiers)
	case C.WIDTH:
		resized = collection.GetWidth(request.name, request.dimension, request.Modifiers)
	case C.HEIGHT:
		resized = collection.GetHeight(request.name, request.dimension, request.Modifiers)
	}
	original := collection.GetOriginal(request.name)

	file := h.NewImage(resized, original, request.force)

	capImage(w, request, file, Conf.MaxWidth, Conf.MaxHeight)
}

func capImage(w http.ResponseWriter, request capRequest, file ImageFile, maxWidth int, maxHeight int) {
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
			0, 13, (C.double)(request.blur), (C.int)(maxWidth), (C.int)(maxHeight))
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
