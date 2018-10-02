package main

func main() {
	m := &Benchy{
		Name:     "asdfasdfasdfasdf",
		BirthDay: 1234567801,
		Phone:    "123-456-7890",
		Siblings: 66,
		Spouse:   1,
		Money:    123.456,
	}
	print(m)
	// stub so gopherjs stops whining
}

type Benchy struct {
	Name     string
	BirthDay int64
	Phone    string
	Siblings int32
	Spouse   byte
	Money    float64
}

type Features struct {
	Dynd          MyInterface
	Bin           []byte
	OtherFeatures []*Features
	DatBenchy     Benchy
	EnumyV        Enumy
}

type Enumy int

const (
	A Enumy = iota
	B
	C
)

func (f *Features) Stuff() {
	// and things
}

type MyInterface interface {
	Stuff()
}
