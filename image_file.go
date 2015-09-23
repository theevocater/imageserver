package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

type image_file interface {
	read() ([]byte, bool)
	write([]byte)
}

type disk_image struct {
	// full filename on disk
	resized_name  string
	original_name string
}

func NewDiskImage(prefix, bucket, dimensions, filename string) *disk_image {
	img := new(disk_image)
	img.original_name = path.Clean(path.Join(prefix, bucket, filename))
	img.resized_name = path.Clean(path.Join(prefix, bucket, dimensions, filename))
	return img
}

func (image *disk_image) read() ([]byte, bool) {
	var file, err = ioutil.ReadFile(image.resized_name)
	var resize = false

	if err != nil {
		file, err = ioutil.ReadFile(image.original_name)
		resize = true
		log.Printf("found original file %s", image.original_name)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("found resized file %s", image.resized_name)
	}
	return file, resize
}

func (image *disk_image) write(resized_image []byte) {
	err := os.MkdirAll(path.Dir(image.resized_name), 0755)
	err = ioutil.WriteFile(image.resized_name, resized_image, 0644)
	if err != nil {
		log.Print(err)
	}
}

type s3_image struct {
	filename string
	bucket   string
	//other stuff???
}

func (image s3_image) read() ([]byte, bool) {
	//todo
	panic("Haven't implemented yet")
}

func (image s3_image) write(resized_image []byte) {
	panic("Haven't implemented yet")
}
