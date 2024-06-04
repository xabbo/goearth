package goearth

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/exp/slices"
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

func zeroOneChr(b bool) byte {
	if b {
		return '1'
	}
	return '0'
}

func makePacket(client ClientType, header *NamedHeader, values ...any) *Packet {
	packet := NewPacket(header)
	packet.Client = client
	packet.Write(values...)
	return packet
}

func resetPos(e *InterceptArgs) {
	e.Packet.Pos = 0
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

	initialized  InitEvent
	activated    VoidEvent
	connected    ConnectEvent
	disconnected VoidEvent

	globalInterceptLock  sync.Mutex
	globalIntercept      InterceptEvent
	interceptsLock       sync.Mutex
	intercepts           map[Header][]*interceptGroup
	persistentIntercepts map[*interceptGroup]struct{}
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
		conn:                 conn,
		headers:              NewHeaders(),
		info:                 info,
		globalIntercept:      InterceptEvent{setup: resetPos},
		persistentIntercepts: map[*interceptGroup]struct{}{},
		intercepts:           map[Header][]*interceptGroup{},
	}
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

// Creates a new packet with the specified message identifier and writes the specified values.
func (ext *Ext) NewPacket(identifier Identifier, values ...any) *Packet {
	packet := &Packet{
		Client: ext.client.Type,
		Header: ext.mustResolveIdentifier(identifier),
	}
	packet.Write(values...)
	return packet
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

// Configures a new intercept builder with the specified identifiers.
func (ext *Ext) Intercept(identifiers ...Identifier) InterceptBuilder {
	set := make(map[Identifier]struct{})
	for _, identifier := range identifiers {
		set[identifier] = struct{}{}
	}

	return &interceptBuilder{
		ext:         ext,
		identifiers: set,
	}
}

// Registers an event handler that is invoked when the game connection is lost.
func (ext *Ext) Disconnected(handler VoidHandler) {
	ext.disconnected.Register(handler)
}

// Sends a packet with the specified message identifier and values to the server or client, based on the identifier direction.
func (ext *Ext) Send(identifier Identifier, values ...any) {
	header := ext.mustResolveIdentifier(identifier)
	packet := makePacket(ext.client.Type, header, values...)
	ext.SendPacket(packet)
}

// Sends the specified packet to the server or client, based on the header direction.
func (ext *Ext) SendPacket(packet *Packet) {
	switch packet.Header.Dir {
	case In, Out:
	default:
		panic(fmt.Errorf("no direction specified on packet header: %+v", packet.Header))
	}
	ext.sendRaw(wrapPacket(packet))
}

// Configures a new inline interceptor targeting the specified message identifiers.
func (e *Ext) Recv(identifiers ...Identifier) InlineInterceptor {
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan *Packet, 1)
	context.AfterFunc(ctx, func() { close(result) })
	intercept := &inlineInterceptor{
		ext:         e,
		identifiers: identifiers,
		result:      result,
		timeout:     time.AfterFunc(time.Minute, cancel),
		ctx:         ctx,
		cancel:      cancel,
	}
	return intercept
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

	var err error
	for {
		_, err = io.ReadFull(ext.conn, buf[:4])
		if err != nil {
			break
		}

		packetLength := int(binary.BigEndian.Uint32(buf[0:4]))
		if packetLength < 2 || packetLength > len(buf) {
			panic(fmt.Errorf("received invalid packet length: %d", packetLength))
		}

		_, err := io.ReadFull(ext.conn, buf[:packetLength])
		if err != nil {
			break
		}

		hdr := outHeader(binary.BigEndian.Uint16(buf[0:2]))
		pkt := Packet{Header: hdr, Data: buf[2:packetLength]}
		switch pkt.Header.Value {
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

	if !errors.Is(err, io.EOF) {
		panic(err)
	}
}

func (ext *Ext) registerInterceptGroup(group *interceptGroup, transient bool, sync bool) {
	if sync {
		ext.interceptsLock.Lock()
		defer ext.interceptsLock.Unlock()
	}
	if ext.isPacketInfoAvailable || transient {
		// resolve all identifiers
		headers := map[Identifier]Header{}
		for identifier := range group.identifiers {
			headers[identifier] = ext.mustResolveIdentifier(identifier).Header
		}
		for _, header := range headers {
			ext.intercepts[header] = append(ext.intercepts[header], group)
		}
	}
	if !transient {
		ext.persistentIntercepts[group] = struct{}{}
	}
}

func wrapPacket(packet *Packet) *Packet {
	pkt := NewPacket(outHeader(gOutSendMessage))
	if packet.Header.Dir == Out {
		pkt.WriteByte(1)
	} else {
		pkt.WriteByte(0)
	}
	pkt.WriteInt(6 + packet.Length())
	pkt.WriteInt(2 + packet.Length())
	pkt.WriteShort(int16(packet.Header.Value))
	pkt.WriteBytes(packet.Data)
	return pkt
}

func (ext *Ext) mustResolveIdentifier(identifier Identifier) *NamedHeader {
	return ext.mustResolve(identifier.Dir, identifier.Name)
}

func (ext *Ext) mustResolve(dir Direction, name string) *NamedHeader {
	switch dir {
	case In, Out:
	default:
		panic("direction must be In or Out")
	}
	header := ext.headers.ByName(dir, name)
	if header == nil {
		panic(fmt.Errorf("failed to resolve %s header: %q", dir.String(), name))
	}
	return header
}

// Acquires the interceptor lock, then registers persistent intercepts
func (ext *Ext) flushInterceptors() {
	ext.interceptsLock.Lock()
	defer ext.interceptsLock.Unlock()

	for intercept := range ext.persistentIntercepts {
		ext.registerInterceptGroup(intercept, true, false)
	}
}

func (ext *Ext) sendRaw(p *Packet) {
	buf := [6]byte{}
	binary.BigEndian.PutUint32(buf[0:], uint32(2+p.Length()))
	binary.BigEndian.PutUint16(buf[4:], p.Header.Value)
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
			dir = Out
		} else {
			dir = In
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
	ext.clearIntercepts()
	ext.disconnected.Dispatch()
}

func (ext *Ext) clearIntercepts() {
	ext.interceptsLock.Lock()
	defer ext.interceptsLock.Unlock()
	for msg := range ext.intercepts {
		delete(ext.intercepts, msg)
	}
}

func dispatchIntercept(handlers []InterceptHandler, index int, keep *int,
	header *NamedHeader, intercept *InterceptArgs) {
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
	intercept.Packet.Pos = 0
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

	dir := In
	if outgoing {
		dir = Out
	}

	header := ext.headers.ByValue(dir, headerValue)
	if header == nil {
		header = &NamedHeader{Header{dir, headerValue}, "?"}
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
		block:  blocked,
		Packet: pktModified,
	}

	ext.dispatchGlobalIntercepts(intercept)
	ext.dispatchIntercepts(intercept)

	// Update the original packet with modified values.
	diff := pktModified.Length() - preLen
	newLen := p.Length() + diff
	p.Header.Value = gOutManipulatedPacket
	p.WriteIntAt(newLen-4, 0)
	p.WriteByteAt(zeroOneChr(intercept.block), 4)
	p.WriteByteAt(zeroOneChr(modified), tabs[2]+1)
	p.WriteIntAt(2+pktModified.Length(), tabs[2]+2)
	p.WriteShortAt(int16(pktModified.Header.Value), tabs[2]+6)
	p.WriteBytesAt(pktModified.Data, tabs[2]+8)
	p.Data = p.Data[:newLen]
	ext.sendRaw(p)
}

func (ext *Ext) dispatchGlobalIntercepts(args *InterceptArgs) {
	ext.globalInterceptLock.Lock()
	defer ext.globalInterceptLock.Unlock()

	keep := 0
	handlers := ext.globalIntercept.handlers
	for i := range ext.globalIntercept.handlers {
		dispatchIntercept(handlers, i, &keep, nil, args)
	}
	ext.globalIntercept.handlers = handlers[:keep]
}

func (ext *Ext) dispatchInterceptGroup(intercept *interceptGroup, args *InterceptArgs) {
	if intercept.dereg {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			panic(fmt.Errorf("error in intercept for %s %q: %s",
				args.Packet.Header.Dir.String(), args.Packet.Header.Name, err))
		}
	}()

	args.Packet.Pos = 0
	intercept.handler(args)
}

func (ext *Ext) snapshotIntercepts(header Header) (snapshot []*interceptGroup, exists bool) {
	ext.interceptsLock.Lock()
	defer ext.interceptsLock.Unlock()

	src, exists := ext.intercepts[header]
	if exists {
		snapshot = make([]*interceptGroup, len(src))
		copy(snapshot, src)
	}
	return
}

func (ext *Ext) dispatchIntercepts(args *InterceptArgs) {
	removals := []*interceptGroup{}

	header := args.Packet.Header.Header
	if intercepts, exist := ext.snapshotIntercepts(header); exist {
		for _, intercept := range intercepts {
			ext.dispatchInterceptGroup(intercept, args)
			if args.dereg {
				intercept.dereg = true
				args.dereg = false
			}
			if intercept.dereg {
				removals = append(removals, intercept)
			}
		}
	}

	if len(removals) > 0 {
		ext.removeIntercepts(removals...)
	}
}

func (ext *Ext) removeIntercepts(intercepts ...*interceptGroup) {
	ext.interceptsLock.Lock()
	defer ext.interceptsLock.Unlock()

	for _, intercept := range intercepts {
		intercept.dereg = true
		_, exist := ext.persistentIntercepts[intercept]
		if !exist {
			continue
		}
		delete(ext.persistentIntercepts, intercept)
		for identifier := range intercept.identifiers {
			header := ext.headers.Get(identifier)
			if header == nil {
				continue
			}
			intercepts := ext.intercepts[header.Header]
			i := slices.Index(intercepts, intercept)
			if i != -1 {
				ext.intercepts[header.Header] = slices.Delete(intercepts, i, i+1)
			}
		}
	}
}
