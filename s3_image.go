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
}

func NewS3Image(s3 *s3gof3r.S3, bucket *s3gof3r.Bucket, resized, original string) ImageFile {
	img := &S3Image{
		s3:           s3,
		bucket:       bucket,
		originalName: original,
		resizedName:  resized,
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

func (image *S3Image) Read() ([]byte, bool) {
	log.Print("trying resized ", image.resizedName)
	if data := readS3(image, image.resizedName); data != nil {
		log.Print("found resized")
		return data, false
	} else {
		log.Print("trying original ", image.originalName)
		return readS3(image, image.originalName), true
	}
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
