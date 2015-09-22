package main

import (
  "io/ioutil"
)

type image_file interface {
  read() []byte
  write([]byte)
}

type disk_image struct {
  filename string
  bucket string
}

func (image disk_image) read() []byte {
  file, err := ioutil.ReadFile(image.filename)
  if err != nil {
    panic(err)
  }
  return file
}

func (image disk_image) write(resized_image []byte) {
  err := ioutil.WriteFile("rerere"+image.filename, resized_image, 0644)
  if err != nil {
    panic(err)
  }
}

type s3_image struct {
  filename string
  bucket string
  //other stuff???
}

func (image s3_image) read() []byte {
  //todo
  panic("Haven't implemented yet")
}

func (image s3_image) write(resized_image []byte) {
  panic("Haven't implemented yet")
}
