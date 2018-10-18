# netgen #
![circleci status](https://circleci.com/gh/lologarithm/netgen.svg?&style=shield)

netgen is a simple go serialization library

Optionally you can generate a copy of the go serializer designed to be run through gopherjs.
C# generator is in progress but has fallen behind with some of the new features and doesn't work at this point.

binary is now located in cmd/netgen

## Usage ##
Package declaration defines the package name of the output (which for now just outputs into the current directory)

Supported field types are:
- Primitives
  - u/int8 (byte)
  - u/int16
  - u/int32
  - u/int64
  - float64
  - string
- Lists
  - Example: []int or []MyStruct
- Structs
- Pointers to Structs
  - Example "MyField \*MyStruct"
  - Primitive pointers are not currently supported
- Enums (always serializes as int32 currently)
- Interfaces (as long as the interface also implements ngen.Net and the structs that implement the interface are defined in the same package)
- Ignored fields using field tag `ngen:"-"`

Use looks like
```
netgen --dir=./my/go/sourcedir --gen=go
```

See the benchmark package for some example generated code.

Currently this generates serialization code for a single package at a time. Imported types will not work.


## Versioned Data ##

Versioning is supported via field tags.

`ngen:X` where X is the order of the field. This allows order of fields to be tracked across different generations of the struct.
Any field without an order number will not be versioned.
A struct can be converted to versioned by setting the field order to be the same as the current struct field order.

If a struct has version tags all fields must be versioned. This is to prevent mistakes in field ordering. Use '-' to ignore a field.

### Example: ###
From:
```
type S struct {
  A int
  B string
  C *S
}
```

To:
```
type S struct {
  A int    `ngen:"2"`
  B string `ngen:"3"`
  C *S     `ngen:"4"`
}
```

Once all producers of the struct are converted to the 'versioned' code you can then start to make changes (removing fields etc)

## Benchmarks ##
These are old benchmarks of the 'unversioned' de/serializers

Benchmarked using go-serialization-benchmark on a lenovo w540 laptop running ubuntu 16.04
```
go version go1.7.1 linux/amd64
BenchmarkMsgpMarshal-8                   	10000000	       201 ns/op	     128 B/op	       1 allocs/op
BenchmarkMsgpUnmarshal-8                 	 3000000	       446 ns/op	     112 B/op	       3 allocs/op
BenchmarkVmihailencoMsgpackMarshal-8     	 1000000	      2174 ns/op	     368 B/op	       6 allocs/op
BenchmarkVmihailencoMsgpackUnmarshal-8   	 1000000	      1978 ns/op	     352 B/op	      13 allocs/op
BenchmarkJsonMarshal-8                   	  300000	      3690 ns/op	    1232 B/op	      10 allocs/op
BenchmarkJsonUnmarshal-8                 	  500000	      3625 ns/op	     416 B/op	       7 allocs/op
BenchmarkEasyJsonMarshal-8               	 1000000	      1501 ns/op	     784 B/op	       5 allocs/op
BenchmarkEasyJsonUnmarshal-8             	 1000000	      1541 ns/op	     160 B/op	       4 allocs/op
BenchmarkBsonMarshal-8                   	 1000000	      1624 ns/op	     392 B/op	      10 allocs/op
BenchmarkBsonUnmarshal-8                 	 1000000	      2176 ns/op	     248 B/op	      21 allocs/op
BenchmarkGobMarshal-8                    	 1000000	      1111 ns/op	      48 B/op	       2 allocs/op
BenchmarkGobUnmarshal-8                  	 1000000	      1102 ns/op	     112 B/op	       3 allocs/op
BenchmarkXdrMarshal-8                    	 1000000	      1929 ns/op	     455 B/op	      20 allocs/op
BenchmarkXdrUnmarshal-8                  	 1000000	      1643 ns/op	     239 B/op	      11 allocs/op
BenchmarkUgorjiCodecMsgpackMarshal-8     	  500000	      2923 ns/op	    2753 B/op	       8 allocs/op
BenchmarkUgorjiCodecMsgpackUnmarshal-8   	  500000	      3252 ns/op	    3008 B/op	       6 allocs/op
BenchmarkUgorjiCodecBincMarshal-8        	  500000	      2967 ns/op	    2785 B/op	       8 allocs/op
BenchmarkUgorjiCodecBincUnmarshal-8      	  500000	      3418 ns/op	    3168 B/op	       9 allocs/op
BenchmarkSerealMarshal-8                 	  500000	      3378 ns/op	     912 B/op	      21 allocs/op
BenchmarkSerealUnmarshal-8               	  500000	      3515 ns/op	    1008 B/op	      34 allocs/op
BenchmarkBinaryMarshal-8                 	 1000000	      1508 ns/op	     256 B/op	      16 allocs/op
BenchmarkBinaryUnmarshal-8               	 1000000	      1630 ns/op	     336 B/op	      22 allocs/op
BenchmarkFlatBuffersMarshal-8            	 5000000	       358 ns/op	       0 B/op	       0 allocs/op
BenchmarkFlatBuffersUnmarshal-8          	 5000000	       299 ns/op	     112 B/op	       3 allocs/op
BenchmarkCapNProtoMarshal-8              	 3000000	       514 ns/op	      56 B/op	       2 allocs/op
BenchmarkCapNProtoUnmarshal-8            	 3000000	       850 ns/op	     200 B/op	       6 allocs/op
BenchmarkCapNProto2Marshal-8             	 1000000	      1238 ns/op	     244 B/op	       3 allocs/op
BenchmarkCapNProto2Unmarshal-8           	 1000000	      1193 ns/op	     320 B/op	       6 allocs/op
BenchmarkHproseMarshal-8                 	 1000000	      1037 ns/op	     473 B/op	       8 allocs/op
BenchmarkHproseUnmarshal-8               	 1000000	      1221 ns/op	     320 B/op	      10 allocs/op
BenchmarkProtobufMarshal-8               	 1000000	      1158 ns/op	     200 B/op	       7 allocs/op
BenchmarkProtobufUnmarshal-8             	 2000000	       759 ns/op	     192 B/op	      10 allocs/op
BenchmarkGoprotobufMarshal-8             	 3000000	       604 ns/op	     312 B/op	       4 allocs/op
BenchmarkGoprotobufUnmarshal-8           	 2000000	       825 ns/op	     432 B/op	       9 allocs/op
BenchmarkGogoprotobufMarshal-8           	10000000	       166 ns/op	      64 B/op	       1 allocs/op
BenchmarkGogoprotobufUnmarshal-8         	 5000000	       265 ns/op	      96 B/op	       3 allocs/op
BenchmarkColferMarshal-8                 	10000000	       144 ns/op	      64 B/op	       1 allocs/op
BenchmarkColferUnmarshal-8               	10000000	       223 ns/op	     112 B/op	       3 allocs/op
BenchmarkGencodeMarshal-8                	10000000	       189 ns/op	      80 B/op	       2 allocs/op
BenchmarkGencodeUnmarshal-8              	10000000	       226 ns/op	     112 B/op	       3 allocs/op
BenchmarkGencodeUnsafeMarshal-8          	20000000	       119 ns/op	      48 B/op	       1 allocs/op
BenchmarkGencodeUnsafeUnmarshal-8        	10000000	       191 ns/op	      96 B/op	       3 allocs/op
BenchmarkNetGenUnmarshal-8               	10000000	       139 ns/op	      64 B/op	       1 allocs/op
BenchmarkNetGenMarshal-8                 	20000000	       101 ns/op	      32 B/op	       2 allocs/op
BenchmarkXDR2Marshal-8                   	10000000	       185 ns/op	      64 B/op	       1 allocs/op
BenchmarkXDR2Unmarshal-8                 	10000000	       176 ns/op	      32 B/op	       2 allocs/op
```
