package models

import "github.com/lologarithm/netgen/lib/ngen"

type Benchy struct {
	Name     string
	BirthDay int64
	Phone    string
	Siblings int32
	Spouse   byte
	Money    float64
}

type FeaturesOne struct {
	Dynd *FeaturesOne
	V    int
}

type Features struct {
	Dynd          MyInterface
	Bin           []byte
	OtherFeatures []*Features
	DatBenchy     Benchy
	EnumyV        Enumy
}

type Enumy int32

const (
	A Enumy = iota
	B
	C
)

func (f *Features) Stuff() {
	// and things
}

type MyInterface interface {
	ngen.Message
	Stuff()
}
