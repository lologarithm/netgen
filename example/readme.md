Example App
----

Running example websocket application first generate network serializers and then run the client and server.

Requires Go 1.13 minimum because syscall/js contains new functions used by this code.

**To Setup:**

1. Generate network files for client `netgen --dir=./example/newmodels/ --gen=go`
2. Generate network files for server `netgen --dir=./example/models/ --gen=go`
3. Launch server with `go run ./example/server/`

**To Run Webassembly Client:**

1. Build client with 'GOARCH=wasm GOOS=js go build -o ./example/client/app/client.wasm ./example/client'
2. Host client with `go run ./example/client/app/host.go --dir=./example/client/app`
3. Open browser to localhost:8080
4. See "Hello World" printed in client console and server log output.

**To Run Native Client:**
1. `go run ./example/app`
2. See messages exchanged between client and server
