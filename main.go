package main

import (
  "net/http"
  "unsafe"
  "regexp"
  "log"
  "os"
  "os/signal"
  "syscall"

  "strings"
  "strconv"

  "time"
)

/*
#cgo CFLAGS: -I/usr/local/include/GraphicsMagick
#cgo LDFLAGS: -L/usr/local/lib -l GraphicsMagick

#include "resize.h"
#include <magick/api.h>
*/
import "C"

// /img/user/256x256/10029684-G5FUMQAZ10SCHGWV.jpg
// /img/collection/dimensions/name?flags
var handler_regxp = regexp.MustCompile("/img/([^/]+)/([^/]+)/([^/]+)")

type Dimensions struct {
  width, height int
}

func Write400(w http.ResponseWriter) {
  if r := recover() ; r != nil {
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

func ParseURL(url string) (string, Dimensions, string) {
  matches := handler_regxp.FindStringSubmatch(url)
  if len(matches) != 4 {
    panic("Couldn't match url")
  }
  collection := matches[1]
  dimensions := ParseDimensions(matches[2])
  name := matches[3]
  return collection, dimensions, name
}

func handler(w http.ResponseWriter, r *http.Request) {
  defer Write400(w)

  bucket, dimensions, name := ParseURL(r.URL.Path)
  file_struct := disk_image{name, bucket}
  file := file_struct.read()

  var length C.size_t = (C.size_t)(len(file))
  blob := C.resize_image(unsafe.Pointer(&file[0]), (*C.size_t)(unsafe.Pointer(&length)), (C.int)(dimensions.width), (C.int)(dimensions.height), 0, 13, 1.0)
  defer C.free(blob)
  // copy to go
  resized_bytes := C.GoBytes(blob, (C.int)(length))
  // is this safe?? I don't want to make too many copies
  go file_struct.write(resized_bytes)

  w.Header().Add("Content-Type", "image/png")
  w.Header().Add("Content-Length", strconv.Itoa((int)(length)))
  w.Header().Add("Last-Modified", time.Now().Format(time.RFC1123))
  w.WriteHeader(http.StatusOK)
  w.Write(resized_bytes)
  w.Write(file)
}

func main() {
  C.init()
  signal_chan := make(chan os.Signal, 1)
  signal.Notify(signal_chan, os.Interrupt)
  signal.Notify(signal_chan, syscall.SIGTERM)
  signal.Notify(signal_chan, syscall.SIGINT)
  go func() {
    <-signal_chan
    C.destroy()
  }()
  http.HandleFunc("/img/", handler)
  log.Print("Starting imageservice")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
