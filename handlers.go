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
		log.Print("Failed to process ", r)
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

const DEFAULTBLUR float64 = 1.0

type Modifiers struct {
	force bool
	blur  float64
}

func NewModifiers() Modifiers {
	return Modifiers{
		force: false,
		blur:  DEFAULTBLUR,
	}
}

func parseModifiers(vars map[string]string, query map[string][]string) Modifiers {
	ret := NewModifiers()
	var err error
	if arr := query["force"]; len(arr) > 0 {
		ret.force, err = strconv.ParseBool(arr[0])
		if err != nil {
			ret.force = false
		}
	}

	if vars["blur"] != "" {
		ret.blur, err = strconv.ParseFloat(vars["blur"], 64)
		if err != nil {
			ret.blur = DEFAULTBLUR
		}
	}

	if arr := query["blur"]; ret.blur != DEFAULTBLUR && len(arr) > 0 {
		ret.blur, err = strconv.ParseFloat(arr[0], 64)
		if err != nil {
			ret.blur = DEFAULTBLUR
		}
	}
	return ret
}
