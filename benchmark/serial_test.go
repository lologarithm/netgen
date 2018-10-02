package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/lologarithm/netgen/lib/ngen"
)

func TestFeaturesOne(t *testing.T) {
	ft := FeaturesOne{
		Dynd: &FeaturesOne{V: 1},
		V:    2,
		// Bin:           []byte{1, 2, 3},
		// OtherFeatures: []*Features{{EnumyV: 11}},
		// EnumyV:        Enumy(10),
	}

	buff := make([]byte, ft.Len())
	ft.Serialize(buff)
	fmt.Printf("buff: %v\n", buff)
	newft := FeaturesOneDeserialize(ngen.NewBuffer(buff))
	if newft.Dynd.V != 1 {
		t.FailNow()
	}
}

func TestFeatures(t *testing.T) {
	ft := Features{
		Dynd:          &Features{},
		Bin:           []byte{1, 2, 3},
		OtherFeatures: []*Features{{EnumyV: 11}},
		EnumyV:        Enumy(10),
	}

	buff := make([]byte, ft.Len())
	ft.Serialize(buff)
	fmt.Printf("buff: %v", buff)
	newft := FeaturesDeserialize(ngen.NewBuffer(buff))

	if len(ft.Bin) != len(newft.Bin) {
		t.Fatalf("Binary blob len doesn't match: %v vs %v", ft.Bin, newft.Bin)
	}
	for i, v := range ft.Bin {
		if newft.Bin[i] != v {
			t.Fatalf("Binary blob doesn't match: %v vs %v", ft.Bin, newft.Bin)
		}
	}

	fmt.Printf("\nnewft: %d\n", newft.EnumyV)
	if newft.EnumyV != 10 {
		t.FailNow()
	}

	if newft.OtherFeatures[0].EnumyV != 11 {
		t.FailNow()
	}
}
func generateNetGen() []*Benchy {
	a := make([]*Benchy, 0, 1000)
	for i := 0; i < 1000; i++ {
		a = append(a, &Benchy{
			Name:     "asdfasdfasdfasdf",
			BirthDay: time.Now().UnixNano(),
			Phone:    "123-456-7890",
			Siblings: rand.Int31n(5),
			Spouse:   byte(rand.Int31n(2)),
			Money:    rand.Float64(),
		})
	}
	return a
}

func BenchmarkNetGenMarshal(b *testing.B) {
	b.StopTimer()
	data := generateNetGen()
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		obj := data[rand.Intn(len(data))]
		obj.Serialize(make([]byte, obj.Len()))
	}
}

func BenchmarkNetGenUnmarshal(b *testing.B) {
	validate := ""
	b.StopTimer()
	data := generateNetGen()
	ser := make([]*ngen.Buffer, len(data))
	for i, d := range data {
		buf := make([]byte, d.Len())
		d.Serialize(buf)
		ser[i] = ngen.NewBuffer(buf)
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n := i % len(ser)
		o := BenchyDeserialize(ser[n])
		ser[n].Reset()
		// Validate unmarshalled data.
		if validate != "" {
			i := data[n]
			correct := o.Name == i.Name && o.Phone == i.Phone && o.Siblings == i.Siblings && o.Spouse == i.Spouse && o.Money == i.Money && o.BirthDay == i.BirthDay
			if !correct {
				b.Fatalf("unmarshaled object differed:\n Expected: %v\n Found: %v", i, o)
			}
		}
	}
}
