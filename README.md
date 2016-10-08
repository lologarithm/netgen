netgen
--------------------

netgen is a simple go serialization library (with the start of C# support)

You first create a netgen definition file (example in here 'defs.ng')

Example
```
package example

struct Heartbeat {
    Time int64
    Latency int64
}

struct Connected {
    Awesomeness Level
}

enum Level {
    PrettyLow = 0
    PrettyOk = 1
    PrettyAwesome = 2
}
```
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
  - Example "MyField *MyStruct"
- Enums

Usage
----------------------
First generate the message definitions
```
netgen mydefs.ng
```

Now in code you could do this:
```
packet, ok := netmsg.NextPacket(buffer) // buffer is assumed to have been read off socket

// You can now check the type
if packet.Frame.MsgType == netmsg.MyMessageType {
    // Process the message!
    ProcessMyMessage(packet.NetMsg.(*netmsg.MyMessage))
}

// Now later you want to respond on the socket.
// First create a packet with your message
packet := netmsg.NewPacket(netmsg.MyResponse, &netmsg.MyResponse{
    AString: "a response message",
    AValue: 100,
})

// To send the message simply
responseBuffer := packet.Pack()
```

By default all messages over the wire are decoded to message pointers.
Fields on these messages could be pointers or structs.
