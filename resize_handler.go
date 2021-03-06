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

type resizeRequest struct {
	width, height int

	collection, name string
	Modifiers
}

func parseResize(vars map[string]string, query map[string][]string) resizeRequest {
	r := resizeRequest{}

	r.collection = vars["collection"]

	var err error

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

	r.Modifiers = parseModifiers(vars, query)

	return r
}

type ResizeHandler struct {
	Handler
}

func (h ResizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we panic, write a 400 and return. this sucks, but uh, its fine for now.
	defer Write400(w)

	request := parseResize(mux.Vars(r), r.URL.Query())

	if request.width > h.MaxWidth || request.height > h.MaxHeight {
		panic("Request image too large")
	}
	collection := h.imageCollections[request.collection]
	resized := collection.GetResized(request.name, request.width, request.height, request.Modifiers)
	original := collection.GetOriginal(request.name)
	file := h.NewImage(resized, original, request.force)
	input, resize := file.Read()

	var output []byte
	var length int

	if resize || request.force {
		length = len(input)
		blob := C.resize_image(unsafe.Pointer(&input[0]), (*C.size_t)(unsafe.Pointer(&length)), (C.int)(request.width), (C.int)(request.height), 0, 13, (C.double)(request.blur))
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
