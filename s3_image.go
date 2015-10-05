package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"

	"github.com/rlmcpherson/s3gof3r"
)

type S3Image struct {
	s3     *s3gof3r.S3
	bucket *s3gof3r.Bucket

	resizedName  string
	originalName string
	force        bool
}

func NewS3Image(s3 *s3gof3r.S3, bucket *s3gof3r.Bucket, resized, original string, force bool) ImageFile {
	img := &S3Image{
		s3:           s3,
		bucket:       bucket,
		originalName: original,
		resizedName:  resized,
		force:        force,
	}
	return img
}

func readS3(image *S3Image, path string) []byte {
	r, _, err := image.bucket.GetReader(path, nil)
	if err != nil {
		return nil
	} else {
		defer r.Close()
	}
	// copy bytes to buffer
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil
	}
	return bytes
}

func (image *S3Image) Read() (data []byte, resize bool) {
	if !image.force {
		data = readS3(image, image.resizedName)
		resize = false
	}

	if data == nil {
		data = readS3(image, image.originalName)
		resize = true
	}
	return data, resize
}

func (image *S3Image) Write(resized_image []byte) error {
	// Open a PutWriter for upload
	w, err := image.bucket.PutWriter(image.resizedName, nil, nil)
	defer w.Close()
	if err != nil {
		log.Print("unable to write to s3")
		return err
	}

	_, err = io.Copy(w, bytes.NewReader(resized_image))
	if err != nil {
		log.Print("unable to write to s3")
		return err
	}
	return nil
}
