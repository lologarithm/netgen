package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/lologarithm/netgen/benchmark/netmsg"
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

func BenchmarkNetGenUnmarshal(b *testing.B) {
	b.StopTimer()
	data := generateNetGen()
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		obj := data[rand.Intn(len(data))]
		obj.Serialize(make([]byte, obj.Len()))
	}
}

func BenchmarkNetGenMarshal(b *testing.B) {
	validate := ""
	b.StopTimer()
	data := generateNetGen()
	ser := make([][]byte, len(data))
	for i, d := range data {
		ser[i] = make([]byte, d.Len())
		d.Serialize(ser[i])
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n := i % len(ser)
		o := netmsg.BenchyDeserialize(ser[n])
		// Validate unmarshalled data.
		if validate != "" {
			i := data[n]
			correct := o.Name == i.Name && o.Phone == i.Phone && o.Siblings == i.Siblings && o.Spouse == i.Spouse && o.Money == i.Money && o.BirthDay == i.BirthDay
			if !correct {
				b.Fatalf("unmarshaled object differed:\n%v\n%v", i, o)
			}
		}
	}
}
