package main

import "fmt"
import "encoding/json"
import "io/ioutil"
import "errors"

type ImageCollection struct {
	Resized  string
	Capped   string
	Height   string
	Width    string
	Original string
}

func ParseImageCollections(file string) (map[string]ImageCollection, error) {
	jsonBlob, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.New("Couldn't read collection data")
	}

	var ret map[string]ImageCollection
	err = json.Unmarshal(jsonBlob, &ret)
	if err != nil {
		return nil, errors.New("Couldn't marshal collection data")
	}

	for k, v := range ret {
		if v.Resized == "" || v.Capped == "" || v.Height == "" || v.Width == "" || v.Original == "" {
			err = errors.New(fmt.Sprintf("Failed to parse %s. Found %s.", k, v))
		}
	}

	return ret, err
}

func addModifiers(mods Modifiers) string {
	ret := ""
	if mods.blur != DEFAULTBLUR {
		ret = fmt.Sprintf("blur%g/", mods.blur)
	}
	return ret
}

func (collection *ImageCollection) GetResized(imageId string, width int, height int, mods Modifiers) string {
	return fmt.Sprintf(collection.Resized, width, height, addModifiers(mods), imageId)
}

func (collection *ImageCollection) GetCapped(imageId string, capped int, mods Modifiers) string {
	return fmt.Sprintf(collection.Capped, capped, addModifiers(mods), imageId)
}

func (collection *ImageCollection) GetWidth(imageId string, width int, mods Modifiers) string {
	return fmt.Sprintf(collection.Width, width, addModifiers(mods), imageId)
}

func (collection *ImageCollection) GetHeight(imageId string, height int, mods Modifiers) string {
	return fmt.Sprintf(collection.Height, height, addModifiers(mods), imageId)
}

func (collection *ImageCollection) GetOriginal(imageId string) string {
	return fmt.Sprintf(collection.Original, imageId)
}
