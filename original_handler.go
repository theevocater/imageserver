package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type request struct {
	collection, name string
}

func parse(vars map[string]string, query map[string][]string) request {
	return request{
		collection: vars["collection"],
		name:       vars["name"],
	}
}

type OriginalHandler struct {
	Handler
}

func (h OriginalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer Write400(w)

	request := parse(mux.Vars(r), r.URL.Query())

	collection := h.imageCollections[request.collection]
	file := h.NewImage("", collection.GetOriginal(request.name), true)
	original, _ := file.Read()

	WriteImage(request.name, original, len(original), w)
}
