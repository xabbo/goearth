package goearth

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

// maximum Habbo packet sizes
const (
	maxIncomingPacketSize uint32 = 1024 * 128
	maxOutgoingPacketSize uint32 = 1024 * 8
)

const (
	gInClick = 1 + iota
	gInInfoRequest
	gInPacketIntercept
	gInFlagsCheck
	gInConnectionStart
	gInConnectionEnd
	gInInit
	gInPacketToStringResponse = 20
	gInStringToPacketResponse = 21
)

const (
	gOutInfo = 1 + iota
	gOutManipulatedPacket
	gOutRequestFlags
	gOutSendMessage
	gOutPacketToStringRequest = 20
	gOutStringToPacketRequest = 21
	gOutExtensionConsoleLog   = 98
)

func zeroOneChr(condition bool) byte {
	if condition {
		return '1'
	}
	return '0'
}

func makePacket(client ClientType, header *Header, values ...any) *Packet {
	packet := NewPacket(header)
	packet.Client = client
	packet.Write(values...)
	return packet
}

// Provides an API to create an extension for G-Earth.
type Ext struct {
	conn      net.Conn
	headers   *Headers
	writeLock sync.Mutex
	info      ExtInfo

	isConnected           bool
	isPacketInfoAvailable bool

	remoteHost string
	remotePort int
	client     Client

	// events

	initialized     InitEvent
	activated       VoidEvent
	connected       ConnectEvent
	disconnected    VoidEvent
	globalIntercept InterceptEvent
	interceptors    map[Direction]map[uint16]*InterceptEvent

	interceptIdentifiers map[Identifier][]InterceptHandler
}

// Defines information about an extension.
type ExtInfo struct {
	Title                string
	Author               string
	Version              string
	Description          string
	ShowEventButton      bool
	IsInstalledExtension bool
	Filename             string
	Cookie               string
	ShowLeaveButton      bool
	ShowDeleteButton     bool
}

// Gets the extension port used to connect to G-Earth.
// Returns -1 if there is no connection.
func (e *Ext) ExtPort() int {
	if e.conn == nil {
		return -1
	}
	switch v := e.conn.RemoteAddr().(type) {
	case *net.TCPAddr:
		return v.Port
	default:
		return -1
	}
}

// Gets the headers used by this extension.
func (ext *Ext) Headers() *Headers {
	return ext.headers
}

// Gets if there is an active connection to the game.
func (e *Ext) IsConnected() bool {
	return e.isConnected
}

// Gets the remote host of the game server.
func (e *Ext) RemoteHost() string {
	return e.remoteHost
}

// Gets the remote port of the game server.
func (e *Ext) RemotePort() int {
	return e.remotePort
}

// Gets the client info for the current connection.
func (ext *Ext) Client() Client {
	return ext.client
}

// Creates a new extension with the provided extension info.
func NewExt(info ExtInfo) *Ext {
	return NewExtWithConn(nil, info)
}

// Creates a new extension with the provided extension info, using the specified connection.
func NewExtWithConn(conn net.Conn, info ExtInfo) *Ext {
	return &Ext{
		conn:    conn,
		headers: NewHeaders(),
		info:    info,
		interceptors: map[Direction]map[uint16]*InterceptEvent{
			INCOMING: {},
			OUTGOING: {},
		},
		interceptIdentifiers: map[Identifier][]InterceptHandler{},
	}
}

// Creates a new packet with the specified direction and message name.
func (ext *Ext) NewPacket(dir Direction, message string) *Packet {
	return &Packet{
		Client: ext.client.Type,
		Header: ext.mustResolve(dir, message),
	}
}

// Creates a new outgoing packet with the specified message name.
func (ext *Ext) OutPacket(message string) *Packet {
	return ext.NewPacket(OUTGOING, message)
}

// Creates a new incoming packet with the specified message name.
func (ext *Ext) InPacket(message string) *Packet {
	return ext.NewPacket(INCOMING, message)
}

// Registers an event handler that is invoked when the extension is initialized by G-Earth.
func (ext *Ext) Initialized(handler InitHandler) {
	ext.initialized.Register(handler)
}

// Registers an event handler that is invoked when the extension is activated by the user.
func (ext *Ext) Activated(handler VoidHandler) {
	ext.activated.Register(handler)
}

// Registers an event handler that is invoked when a game connection is established.
func (ext *Ext) Connected(handler ConnectHandler) {
	ext.connected.Register(handler)
}

// Registers an event handler that is invoked when a packet is intercepted.
func (ext *Ext) InterceptAll(handler InterceptHandler) {
	ext.globalIntercept.Register(handler)
}

func (ext *Ext) Intercept(messages ...string) InterceptBuilder {
	return &interceptBuilder{
		Ext:      ext,
		Messages: messages,
	}
}

// Registers an event handler that is invoked when
// an packet with the specified message name and direction is intercepted.
func (ext *Ext) addIntercept(dir Direction, name string, handler InterceptHandler) {
	ext.interceptIdentifier(Identifier{dir, name}, handler)
}

// Registers an event handler that is invoked when
// an incoming packet with the specified message name is intercepted.
func (ext *Ext) addInterceptIn(name string, handler InterceptHandler) {
	ext.addIntercept(INCOMING, name, handler)
}

// Registers an event handler that is invoked when
// an outgoing packet with the specified message name is intercepted.
func (ext *Ext) addInterceptOut(name string, handler InterceptHandler) {
	ext.addIntercept(OUTGOING, name, handler)
}

// Registers an event handler that is invoked when the game connection is lost.
func (ext *Ext) Disconnected(handler VoidHandler) {
	ext.disconnected.Register(handler)
}

// Sends a packet with the specified message name and values to the server.
func (ext *Ext) Send(name string, values ...any) {
	header := ext.mustResolve(OUTGOING, name)
	packet := makePacket(ext.client.Type, header, values...)
	ext.SendP(packet)
}

// Sends the specified packet to the server.
func (ext *Ext) SendP(packet *Packet) {
	if packet.Header.dir != OUTGOING {
		panic(fmt.Errorf("cannot send incoming message to server"))
	}
	ext.sendRaw(wrapPacket(packet))
}

// Sends a packet with the specified message name and values to the client.
func (ext *Ext) SendToClient(name string, values ...any) {
	header := ext.mustResolve(INCOMING, name)
	packet := makePacket(ext.client.Type, header, values...)
	ext.SendToClientP(packet)
}

// Sends the specified packet to the client.
func (ext *Ext) SendToClientP(packet *Packet) {
	if packet.Header.dir != INCOMING {
		panic(fmt.Errorf("cannot send outgoing message to client"))
	}
	ext.sendRaw(wrapPacket(packet))
}

func (ext *Ext) Connect(port int) error {
	if ext.conn != nil {
		return fmt.Errorf("the extension is already associated with a connection")
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	ext.conn = conn
	return err
}

func (ext *Ext) MustConnect(port int) {
	err := ext.Connect(port)
	if err != nil {
		panic(err)
	}
}

// Runs the extension processing loop.
// If the extension does not have a connection, one will be initiated
// using the port, cookie and filename command-line arguments via the flag package.
// If you do not want this behaviour, you must first call Connect before Run.
// This method will panic if any errors other than io.EOF occur.
func (ext *Ext) Run() {
	if ext.conn == nil {
		args := parseArgs()
		if args.Cookie != "" {
			ext.info.Cookie = args.Cookie
		}
		if args.Filename != "" {
			ext.info.Filename = args.Filename
			ext.info.IsInstalledExtension = true
			ext.info.ShowLeaveButton = true
			ext.info.ShowDeleteButton = true
		}
		ext.info.ShowLeaveButton = true
		ext.MustConnect(args.Port)
	}

	ext.info.ShowEventButton = len(ext.activated.handlers) > 0

	// allocate buffer with extension protocol overhead
	buf := make([]byte, 64+maxIncomingPacketSize)
	var idx int

	for {
		idx = 0
		for idx < 4 {
			n, err := ext.conn.Read(buf[idx:4])
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				panic(err)
			}
			idx += n
		}
		packetLength := int(binary.BigEndian.Uint32(buf[0:4]))
		if packetLength < 2 || packetLength > len(buf) {
			panic(fmt.Errorf("received invalid packet length: %d", packetLength))
		}
		idx = 0
		for idx < packetLength {
			n, err := ext.conn.Read(buf[idx:packetLength])
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				panic(err)
			}
			idx += n
		}
		hdr := outHeader(binary.BigEndian.Uint16(buf[0:2]))
		pkt := Packet{Header: hdr, Data: buf[2:packetLength]}
		switch pkt.Header.value {
		case gInInfoRequest:
			ext.handleInfoRequest(&pkt)
		case gInInit:
			ext.handleInit(&pkt)
		case gInClick:
			ext.handleActivated(&pkt)
		case gInConnectionStart:
			ext.handleConnectionStart(&pkt)
		case gInConnectionEnd:
			ext.handleConnectionEnd(&pkt)
		case gInPacketIntercept:
			ext.handlePacketIntercept(&pkt)
		}
	}
}

func (ext *Ext) interceptIdentifier(identifier Identifier, handler InterceptHandler) {
	if ext.isPacketInfoAvailable {
		h := ext.mustResolveIdentifier(identifier)
		ext.registerIntercept(h.dir, h.value, handler)
	}
	ext.interceptIdentifiers[identifier] = append(ext.interceptIdentifiers[identifier], handler)
}

func (ext *Ext) registerIntercept(dir Direction, value uint16, handler InterceptHandler) {
	// TODO interceptor lock
	event, exist := ext.interceptors[dir][value]
	if !exist {
		event = &InterceptEvent{}
		ext.interceptors[dir][value] = event
	}
	event.Register(handler)
}

func wrapPacket(packet *Packet) *Packet {
	pkt := NewPacket(outHeader(gOutSendMessage))
	if packet.Header.Dir() == OUTGOING {
		pkt.WriteByte(1)
	} else {
		pkt.WriteByte(0)
	}
	pkt.WriteInt(6 + packet.Length())
	pkt.WriteInt(2 + packet.Length())
	pkt.WriteShort(int16(packet.Header.Value()))
	pkt.WriteBytes(packet.Data)
	return pkt
}

func (ext *Ext) mustResolveIdentifier(identifier Identifier) *Header {
	return ext.mustResolve(identifier.Dir, identifier.Name)
}

func (ext *Ext) mustResolve(dir Direction, name string) *Header {
	header := ext.headers.ByName(dir, name)
	if header == nil {
		dirName := "outgoing"
		if dir == INCOMING {
			dirName = "incoming"
		}
		panic(fmt.Errorf("failed to resolve %s header: %q", dirName, name))
	}
	return header
}

func (ext *Ext) flushInterceptors() {
	for id := range ext.interceptIdentifiers {
		header := ext.mustResolveIdentifier(id)
		for _, handler := range ext.interceptIdentifiers[id] {
			ext.registerIntercept(header.dir, header.value, handler)
		}
	}
}

func (ext *Ext) sendRaw(p *Packet) {
	buf := [6]byte{}
	binary.BigEndian.PutUint32(buf[0:], uint32(2+p.Length()))
	binary.BigEndian.PutUint16(buf[4:], p.Header.Value())
	ext.writeLock.Lock()
	defer ext.writeLock.Unlock()
	ext.conn.Write(buf[:])
	ext.conn.Write(p.Data)
}

func (ext *Ext) handleInit(p *Packet) {
	connected := false
	if p.Length() > 0 {
		connected = p.ReadBool()
	}
	ext.initialized.Dispatch(&InitArgs{
		Connected: connected,
	})
}

func (ext *Ext) handleInfoRequest(p *Packet) {
	res := NewPacket(outHeader(gOutInfo))
	res.Write(&ext.info)
	ext.sendRaw(res)
}

func (ext *Ext) handleActivated(p *Packet) {
	ext.activated.Dispatch()
}

func (ext *Ext) handleConnectionStart(p *Packet) {
	args := ConnectArgs{}
	p.Read(&args)

	for _, msg := range args.Messages {
		var dir Direction
		if msg.Outgoing {
			dir = OUTGOING
		} else {
			dir = INCOMING
		}
		ext.headers.Add(NewHeader(dir, uint16(msg.Id), msg.Name))
	}

	ext.isConnected = true
	ext.isPacketInfoAvailable = true
	ext.remoteHost = args.Host
	ext.remotePort = args.Port
	ext.client = args.Client
	ext.flushInterceptors()
	ext.connected.Dispatch(&args)
}

func (ext *Ext) handleConnectionEnd(p *Packet) {
	ext.isConnected = false
	ext.disconnected.Dispatch()
}

func dispatchIntercept(handlers []InterceptHandler, index int, keep *int,
	header *Header, intercept *InterceptArgs) {
	defer func() {
		if err := recover(); err != nil {
			if header == nil {
				panic(fmt.Errorf("error in global intercept handler: %s", err))
			} else {
				panic(fmt.Errorf("error in intercept handler for %+v: %s", header, err))
			}
		}
	}()
	handler := handlers[index]
	handler(intercept)
	if intercept.dereg {
		intercept.dereg = false
	} else {
		handlers[*keep] = handler
		*keep++
	}
}

func (ext *Ext) handlePacketIntercept(p *Packet) {
	tabs := make([]int, 0, 3)
	for i := 4; i < len(p.Data) && len(tabs) < 3; i++ {
		if p.Data[i] == 0x09 {
			tabs = append(tabs, i)
		}
	}
	if len(tabs) != 3 {
		panic("invalid packet intercept data (insufficient delimiter bytes)")
	}

	blocked := p.Data[4] == '1'
	seq, err := strconv.Atoi(string(p.Data[tabs[0]+1 : tabs[1]]))
	if err != nil {
		panic("failed to parse packet index")
	}
	outgoing := p.Data[tabs[1]+3] == 'S'
	modified := p.Data[tabs[2]+1] == '1'
	headerValue := binary.BigEndian.Uint16(p.Data[tabs[2]+6:])
	packetData := p.Data[tabs[2]+8:]
	preLen := len(packetData)

	dir := INCOMING
	if outgoing {
		dir = OUTGOING
	}

	header := ext.headers.ByValue(dir, headerValue)
	if header == nil {
		header = &Header{dir, headerValue, "?"}
	}

	pktModified := &Packet{
		Client: ext.client.Type,
		Header: header,
		Data:   packetData,
	}

	intercept := &InterceptArgs{
		ext:    ext,
		dir:    dir,
		seq:    seq,
		Packet: pktModified,
		Block:  blocked,
	}

	keep := 0
	handlers := ext.globalIntercept.handlers
	for i := range ext.globalIntercept.handlers {
		intercept.Packet.Pos = 0
		dispatchIntercept(handlers, i, &keep, nil, intercept)
	}
	ext.globalIntercept.handlers = handlers[:keep]

	if interceptors, exist := ext.interceptors[dir][headerValue]; exist {
		handlers := interceptors.handlers
		keep := 0
		for i := range handlers {
			intercept.Packet.Pos = 0
			dispatchIntercept(handlers, i, &keep, header, intercept)
		}
		interceptors.handlers = handlers[:keep]
	}

	// Update the original packet with modified values.
	diff := pktModified.Length() - preLen
	newLen := p.Length() + diff
	p.Header.value = gOutManipulatedPacket
	p.WriteIntAt(newLen-4, 0)
	p.WriteByteAt(zeroOneChr(intercept.Block), 4)
	p.WriteByteAt(zeroOneChr(modified), tabs[2]+1)
	p.WriteIntAt(2+pktModified.Length(), tabs[2]+2)
	p.WriteShortAt(int16(pktModified.Header.value), tabs[2]+6)
	p.WriteBytesAt(pktModified.Data, tabs[2]+8)
	p.Data = p.Data[:newLen]
	ext.sendRaw(p)
}
