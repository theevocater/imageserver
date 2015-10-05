package main

import (
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"time"
)

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

type Handler struct {
	Confs
	imageCollections map[string]ImageCollection
	*ImageFactory
}
