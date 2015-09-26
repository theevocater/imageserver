package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

type DiskImage struct {
	// full filename on disk
	resizedName  string
	originalName string
}

func NewDiskImage(resized, original string) ImageFile {
	img := new(DiskImage)
	img.originalName = original
	img.resizedName = resized
	return img
}

func (image *DiskImage) Read() ([]byte, bool) {
	var file, err = ioutil.ReadFile(image.resizedName)
	var resize = false

	if err != nil {
		file, err = ioutil.ReadFile(image.originalName)
		resize = true
		log.Printf("found original file %s", image.originalName)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("found resized file %s", image.resizedName)
	}
	return file, resize
}

func (image *DiskImage) Write(resized_image []byte) error {
	err := os.MkdirAll(path.Dir(image.resizedName), 0755)
	err = ioutil.WriteFile(image.resizedName, resized_image, 0644)
	if err != nil {
		return err
	}
	return nil
}
