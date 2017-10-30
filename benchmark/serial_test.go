package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/lologarithm/netgen/benchmark/netmsg"
	"github.com/lologarithm/netgen/lib/ngen"
)

func generateNetGen() []*netmsg.Benchy {
	a := make([]*netmsg.Benchy, 0, 1000)
	for i := 0; i < 1000; i++ {
		a = append(a, &netmsg.Benchy{
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
		o := netmsg.BenchyDeserialize(ser[n])
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
