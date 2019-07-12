package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/lologarithm/netgen/benchmark/models"
	"github.com/lologarithm/netgen/lib/ngen"
)

func TestFeaturesOne(t *testing.T) {
	val := 1
	ft := models.FeaturesOne{
		Dynd: &models.FeaturesOne{V: val},
		V:    2,
		// Bin:           []byte{1, 2, 3},
		// OtherFeatures: []*Features{{EnumyV: 11}},
		// EnumyV:        Enumy(10),
	}

	buf := ngen.NewBuffer(make([]byte, ft.Length(nil)))
	ft.Serialize(models.Context, buf)
	newft := models.DeserializeFeaturesOne(models.Context, ngen.NewBuffer(buf.Bytes()))
	if newft.Dynd.V != val {
		t.FailNow()
	}
}

func TestFeatures(t *testing.T) {
	ft := models.Features{
		Dynd:          &models.Features{},
		Bin:           []byte{1, 2, 3},
		OtherFeatures: []*models.Features{{EnumyV: 11}},
		EnumyV:        models.Enumy(10),
	}

	buf := ngen.NewBuffer(make([]byte, ft.Length(nil)))
	ft.Serialize(nil, buf)
	newft := models.DeserializeFeatures(nil, ngen.NewBuffer(buf.Bytes()))

	if len(ft.Bin) != len(newft.Bin) {
		t.Fatalf("Binary blob len doesn't match: %v vs %v", ft.Bin, newft.Bin)
	}
	for i, v := range ft.Bin {
		if newft.Bin[i] != v {
			t.Fatalf("Binary blob doesn't match: %v vs %v", ft.Bin, newft.Bin)
		}
	}
	if newft.EnumyV != 10 {
		t.FailNow()
	}

	if newft.OtherFeatures[0].EnumyV != 11 {
		t.FailNow()
	}
}
func generateNetGen() []*models.Benchy {
	a := make([]*models.Benchy, 0, 1000)
	for i := 0; i < 1000; i++ {
		a = append(a, &models.Benchy{
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
		buf := ngen.NewBuffer(make([]byte, obj.Length(nil)))
		obj.Serialize(nil, buf)
	}
}

func BenchmarkNetGenUnmarshal(b *testing.B) {
	validate := ""
	b.StopTimer()
	data := generateNetGen()
	ser := make([]*ngen.Buffer, len(data))
	for i, d := range data {
		buf := ngen.NewBuffer(make([]byte, d.Length(nil)))
		d.Serialize(nil, buf)
		ser[i] = ngen.NewBuffer(buf.Bytes())
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n := i % len(ser)
		o := models.DeserializeBenchy(nil, ser[n])
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
