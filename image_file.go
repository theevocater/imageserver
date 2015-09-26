package main

import (
	"log"

	"github.com/rlmcpherson/s3gof3r"
)

type ImageFile interface {
	Read() ([]byte, bool)
	Write([]byte) error
}

// if i care to, it might be worth adding a Shutdown to close connections etc
// etc
type ImageFactory struct {
	// the 4x string interface is totally wrong, need to rethink it.
	NewImage func(string, string) ImageFile
}

func NewS3ImageFactory(bucketName string) *ImageFactory {
	factory := new(ImageFactory)
	//log.Print(imageCollections)
	k, err := s3gof3r.EnvKeys() // get S3 keys from environment
	if err != nil {
		log.Fatal("Unable to init s3", err)
	}

	// Open bucket to put file into
	s3 := s3gof3r.New("", k)

	bucket := s3.Bucket(bucketName)
	if bucket == nil {
		log.Fatal("Unable to init s3", err)
	}
	// HOLY SMOKES IT DOESN'T WORKS
	factory.NewImage = func(r, o string) ImageFile {
		return NewS3Image(s3, bucket, r, o)
	}
	return factory
}

func NewDiskImageFactory() *ImageFactory {
	factory := new(ImageFactory)
	factory.NewImage = NewDiskImage
	return factory
}
