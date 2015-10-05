package main

import (
	"io/ioutil"
	"os"
	"path"
)

type DiskImage struct {
	// full filename on disk
	resizedName  string
	originalName string
	force        bool
}

func NewDiskImage(resized, original string, force bool) ImageFile {
	img := &DiskImage{
		originalName: original,
		resizedName:  resized,
		force:        force,
	}
	return img
}

func (image *DiskImage) Read() (data []byte, resize bool) {
	var err error

	if !image.force {
		data, err = ioutil.ReadFile(image.resizedName)
		resize = false
	}

	if err != nil || data == nil {
		data, err = ioutil.ReadFile(image.originalName)
		resize = true
		if err != nil {
			data = nil
		}
	}
	return data, resize
}

func (image *DiskImage) Write(resized_image []byte) error {
	err := os.MkdirAll(path.Dir(image.resizedName), 0755)
	err = ioutil.WriteFile(image.resizedName, resized_image, 0644)
	if err != nil {
		return err
	}
	return nil
}
