# goearth
Go extension API for the Habbo packet interceptor [G-Earth](https://github.com/sirjonasxx/G-Earth).

## Requirements

Requires [Go](https://go.dev/dl/) 1.21+.

## Usage

Check out the [examples](.examples/) for reference.

### Getting started

1. Create a new directory and save the following code as `main.go`.

```go
package main

import g "github.com/b7c/goearth"

var ext = g.NewExt(g.ExtInfo{
    Title: "Your extension",
    Description: "Go Earth!",
    Version: "1.0",
    Author: "You",
})

func main() {
    // register event handlers and interceptors here
    ext.Run()
}
```

2. Create a new module and install dependencies:
```sh
go mod init your-extension
go mod tidy
```

3. Run the extension:

```sh
go run .
```

You should see your extension appear in G-Earth's extension list.

Import the `in`/`out` packages to access the respective incoming/outgoing message identifiers.
```go
import "github.com/b7c/goearth/in"
import "github.com/b7c/goearth/out"
```

### Events

#### On extension initialized

```go
ext.Initialized(func(e *g.InitArgs) {
    log.Printf("Extension initialized (connected=%t)", e.Connected)
})
```

#### On game connected

```go
ext.Connected(func(e *g.ConnectArgs) {
    log.Printf("Game connected (%s:%d)", e.Host, e.Port)
    log.Printf("Client %s (%s)", e.Client.Identifier, e.Client.Version)
})
```

#### On extension activated

This is when the extension's green "play" button is clicked in G-Earth.
```go
ext.Activated(func() {
    log.Println("Extension clicked in G-Earth")
})
```

#### On game disconnected

```go
ext.Disconnected(func() {
    log.Println("Game disconnected")    
})
```

### Intercepting packets

#### All packets

```go
ext.InterceptAll(func (e *g.InterceptArgs) {
    log.Printf("Intercepted %s message %q\n",
        e.Dir(), e.Packet.Header.Name)
})
```

#### By name

```go
ext.Intercept(in.Chat, in.Shout, in.Whisper).With(func (e *g.InterceptArgs) {
    idx := e.Packet.ReadInt()
    msg := e.Packet.ReadString()
    log.Printf("Entity #%d said %q", idx, msg)
})
```

#### Blocking packets

```go
ext.Intercept(out.MoveAvatar).With(func (e *g.InterceptArgs) {
    // prevent movement
    e.Block = true
})
```

#### Modifying packets

```go
ext.Intercept(in.Chat, in.Shout).With(func(e *g.InterceptArgs) {
    // make everyone's chat messages uppercase
    e.Packet.ModifyStringAt(strings.ToUpper, 4)
})
```

### Reading packets

#### By type

```go
x := pkt.ReadInt()
y := pkt.ReadInt()
z := pkt.ReadString()
```

#### By pointer

```go
var x, y int
var z string
pkt.Read(&x, &y, &z)
```

#### Into a struct

```go
type MyStruct struct {
    X, Y int
    Z    string
}
```

```go
myStruct := MyStruct{}
pkt.Read(&myStruct)
```

#### Using a custom parser by implementing Parsable

```go
type MyParsable struct {
    X, Y int
    Z    float32 
}

func (v *MyParsable) Parse(p *g.Packet, pos *int) {
    // perform custom parsing logic here
    x := p.ReadIntPtr(pos)
    y := p.ReadIntPtr(pos)
    zStr := p.ReadStringPtr(pos)
    z, err := strconv.ParseFloat(zStr, 32)
    if err != nil {
        panic(err)
    }
    *v = MyParsable{ X: x, Y: y, Z: float32(z) }
}
```

```go
myParsable := MyParsable{}
// myParsable.Parse(...) will be invoked
pkt.Read(&myParsable)
```

### Writing packets

#### By type

```go
x, y := 1, 2
z := "3.0"
pkt.WriteInt(x)
pkt.WriteInt(y)
pkt.WriteString(z)
```

#### By values

```go
x, y := 1, 2
z := "3.0"
pkt.Write(x, y, z)
```

#### By struct

```go
type MyStruct struct {
    X, Y int
    Z    string
}
myStruct := MyStruct{X: 1, Y: 2, Z: "3.0"}
pkt.Write(myStruct)
```

### Sending packets

#### By values

```go
// to server
ext.Send(out.Chat, "hello, world", 0, -1)
// to client
ext.Send(in.Chat, 0, "hello, world", 0, 34, 0, 0)
// take care when sending packets to the client
// as badly formed packets will crash the game client
```

#### By packet

```go
pkt := ext.NewPacket(in.Chat)
pkt.WriteInt(0)
pkt.WriteString("hello, world")
pkt.WriteInt(0)
pkt.WriteInt(34)
pkt.WriteInt(0)
pkt.WriteInt(0)
ext.SendPacket(pkt)
```

### Receiving packets

```go
log.Println("Retrieving user info...")
ext.Send(out.InfoRetrieve)
if pkt := ext.Recv(in.UserObject).Wait(); pkt != nil {
    id, name := pkt.ReadInt(), pkt.ReadString()
    log.Printf("Got user info (id: %d, name: %q)", id, name)
} else {
    log.Println("Timed out")
}
```

Note that calling `Wait()` from an event handler would cause the extension to hang as event handlers are invoked from the main packet processing loop.\
Launch a goroutine if you need to do any inline receiving of packets, for example:

```go
ext.Activated(func() {
    go doStuff()
})
```

```go
func doStuff() {
    // send and receive packets here 
}
```
