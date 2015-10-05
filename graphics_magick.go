package main

/*
#cgo pkg-config: GraphicsMagick

#include "resize.h"
*/
import "C"

func InitMagick() {
	C.InitializeMagick(nil)
}

func CloseMagick() {
	C.DestroyMagick()
}
