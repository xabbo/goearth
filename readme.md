# goearth
Go extension API for the Habbo packet interceptor [G-Earth](https://github.com/sirjonasxx/G-Earth).

## Requirements

Requires [Go](https://go.dev/dl/) 1.22+.

## Usage

Check out the [examples](.examples/) for reference.

### Getting started

1. Create a new directory and save the following code as `main.go`.

```go
package main

import g "xabbo.b7c.io/goearth"

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
import "xabbo.b7c.io/goearth/in"
import "xabbo.b7c.io/goearth/out"
```

For the Shockwave messages, use the `xabbo.b7c.io/goearth/shockwave/in` and `out` packages.

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
ext.InterceptAll(func (e *g.Intercept) {
    log.Printf("Intercepted %s message %q\n", e.Dir(), e.Name())
})
```

#### By name

```go
ext.Intercept(in.Chat, in.Shout, in.Whisper).With(func (e *g.Intercept) {
    idx := e.Packet.ReadInt()
    msg := e.Packet.ReadString()
    log.Printf("Entity #%d said %q", idx, msg)
})
```

#### Blocking packets

```go
ext.Intercept(out.MoveAvatar).With(func (e *g.Intercept) {
    // prevent movement
    e.Block()
})
```

#### Modifying packets

```go
ext.Intercept(in.Chat, in.Shout).With(func(e *g.Intercept) {
    // make everyone's chat messages uppercase
    e.Packet.ModifyStringAt(4, strings.ToUpper)
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
type Tile struct {
    X, Y int
    Z    float32
}
```

```go
tile := Tile{}
pkt.Read(&tile)
```

#### Using a custom parser by implementing Parsable

```go
type Tile struct {
    X, Y int
    Z    float32 
}

func (v *Tile) Parse(p *g.Packet, pos *int) {
    // perform custom parsing logic here
    // make sure to use the Read*Ptr variants here
    // to ensure the provided position is advanced properly
    x := p.ReadIntPtr(pos)
    y := p.ReadIntPtr(pos)
    zStr := p.ReadStringPtr(pos)
    z, err := strconv.ParseFloat(zStr, 32)
    if err != nil {
        panic(err)
    }
    *v = Tile{ X: x, Y: y, Z: float32(z) }
}
```

```go
tile := Tile{}
// Tile.Parse(...) will be invoked
pkt.Read(&tile)
```

### Writing packets

#### By type

```go
pkt.WriteInt(1)
pkt.WriteInt(2)
pkt.WriteString("3.0")
```

#### By values

```go
// writes int, int, string
pkt.Write(1, 2, "3.0")
```

#### By struct

```go
type Tile struct {
    X, Y int
    Z    float32
}
tile := Tile{X: 1, Y: 2, Z: 3.0}
pkt.Write(tile)
```

#### Using a custom composer by implementing Composable

```go
type Tile struct {
    X, Y int
    Z    float32 
}

func (v Tile) Compose(p *g.Packet, pos *int) {
    // perform custom composing logic here
    // make sure to use the Write*Ptr variants here
    // to ensure the provided position is advanced properly
    p.WriteIntPtr(pos, v.X)
    p.WriteIntPtr(pos, v.Y)
    p.WriteStringPtr(pos, strconv.FormatFloat(v.Z, 'f', -1, 32))
}
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

**Note:** do not perform any long running operations inside an intercept handler.\
If you attempt to `Wait` for a packet inside an intercept handler,\
you will never receive it as the packet's processing loop will be paused until it times out.\
Launch a goroutine with the `go` keyword if you need to do this inside an intercept handler, for example:

```go
ext.Intercept(in.Chat).With(func(e *g.Intercept) {
    // also, do not pass Packets to another goroutine as its buffer may no longer be valid.
    // read any values within the intercept handler and then pass those.
    msg := e.Packet.ReadStringAt(4)
    go func() {
        // perform long running operation here...
        time.Sleep(10 * time.Second)
        ext.Send(out.Shout, msg)
    }()
})
```
