Example App
----

Running example websocket application first generate network serializers and then run the client and server.

**To Setup:**

1. Generate network files for client `netgen --dir=./example/newmodels/ --gen=go,js`
2. Generate network files for server `netgen --dir=./example/models/ --gen=go`

**To Run:**

1. Launch server with `go run ./example/server/`
2. Host client with `gopherjs serve github.com/lologarithm/netgen/example/client`
3. Open browser to localhost:8080
4. See "Hello World" printed in client console and server log output.
