package goearth

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"maps"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	"xabbo.b7c.io/goearth/encoding"
	"xabbo.b7c.io/goearth/internal/debug"
)

var dbgExt = debug.NewLogger("[ext]")

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

func resetPos(e *Intercept) {
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

	connectionCtx      context.Context
	closeConnectionCtx context.CancelFunc

	// events

	initialized  InitEvent
	activated    VoidEvent
	connected    ConnectEvent
	disconnected VoidEvent

	globalInterceptLock  sync.Mutex
	globalIntercept      InterceptEvent
	interceptsLock       sync.Mutex
	intercepts           map[Header][]*interceptRegistration
	persistentIntercepts map[*interceptRegistration]struct{}
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

func (e *Ext) Context() context.Context {
	return e.connectionCtx
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
		persistentIntercepts: map[*interceptRegistration]struct{}{},
		intercepts:           map[Header][]*interceptRegistration{},
	}
}

func (ext *Ext) Connect(port int) error {
	host, ok := os.LookupEnv("GOEARTH_HOST")
	if !ok {
		host = "127.0.0.1"
	}
	if ext.conn != nil {
		return fmt.Errorf("the extension is already associated with a connection")
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	ext.conn = conn
	return err
}

func (ext *Ext) Log(a ...any) {
	p := &Packet{Header: Header{Out, gOutExtensionConsoleLog}}
	p.WriteString(ext.info.Title + " --> " + fmt.Sprint(a...))
	ext.sendRaw(p)
}

func (ext *Ext) Logf(format string, a ...any) {
	p := &Packet{Header: Header{Out, gOutExtensionConsoleLog}}
	p.WriteString(ext.info.Title + " --> " + fmt.Sprintf(format, a...))
	ext.sendRaw(p)
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
		ix:          ext,
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
	packet := &Packet{
		Client: ext.client.Type,
		Header: header,
	}
	packet.Write(values...)
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
		ix:          e,
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
	if err := ext.RunE(); err != nil {
		panic(err)
	}
}

// Runs the extension processing loop.
// If the extension does not have a connection, one will be initiated
// using the port, cookie and filename command-line arguments via the flag package.
// If you do not want this behaviour, you must first call Connect before Run.
func (ext *Ext) RunE() (err error) {
	defer func() {
		if e := recover(); e != nil {
			if recoveredErr, ok := e.(error); ok {
				err = fmt.Errorf("unhandled error: %w", recoveredErr)
			} else {
				err = fmt.Errorf("unhandled error: %s", recoveredErr)
			}
		}
	}()

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
		err = ext.Connect(args.Port)
		if err != nil {
			return
		}
	}

	defer func() {
		if conn := ext.conn; conn != nil {
			ext.conn.Close()
		}
	}()

	ext.info.ShowEventButton = len(ext.activated.handlers) > 0

	// allocate buffer with extension protocol overhead
	buf := make([]byte, 64+maxIncomingPacketSize)

	for err == nil {
		_, err = io.ReadFull(ext.conn, buf[:4])
		if err != nil {
			break
		}

		packetLength := int(binary.BigEndian.Uint32(buf[0:4]))
		if packetLength < 2 || packetLength > len(buf) {
			err = fmt.Errorf("received invalid packet length: %d", packetLength)
			return
		}

		_, err = io.ReadFull(ext.conn, buf[:packetLength])
		if err != nil {
			break
		}

		pkt := Packet{
			Header: Header{Out, binary.BigEndian.Uint16(buf[0:2])},
			Data:   buf[2:packetLength],
		}

		switch pkt.Header.Value {
		case gInInfoRequest:
			ext.handleInfoRequest()
		case gInInit:
			ext.handleInit(&pkt)
		case gInClick:
			ext.handleActivated()
		case gInConnectionStart:
			ext.handleConnectionStart(&pkt)
		case gInConnectionEnd:
			ext.handleConnectionEnd()
		case gInPacketIntercept:
			err = ext.handlePacketIntercept(&pkt)
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

func (ext *Ext) Register(group *InterceptGroup) InterceptRef {
	reg := &interceptRegistration{
		ext:         ext,
		identifiers: maps.Clone(group.Identifiers),
		handler:     group.Handler,
	}
	ext.registerInterceptGroup(reg, group.Transient, true)
	return reg
}

func (ext *Ext) registerInterceptGroup(group *interceptRegistration, transient bool, sync bool) {
	if sync {
		ext.interceptsLock.Lock()
		defer ext.interceptsLock.Unlock()
	}
	if ext.isPacketInfoAvailable || transient {
		// resolve all identifiers
		headers := map[Identifier]Header{}
		for identifier := range group.identifiers {
			headers[identifier] = ext.mustResolveIdentifier(identifier)
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
	pkt := &Packet{Header: Header{Out, gOutSendMessage}} // NewPacket(outHeader(gOutSendMessage))
	if packet.Header.Dir == Out {
		pkt.WriteByte(1)
	} else {
		pkt.WriteByte(0)
	}
	if packet.Client != Shockwave {
		pkt.WriteInt(6 + packet.Length())
	}
	pkt.WriteInt(2 + packet.Length())
	if packet.Client == Shockwave {
		B64(packet.Header.Value).Compose(pkt, &pkt.Pos)
	} else {
		pkt.WriteShort(int16(packet.Header.Value))
	}
	pkt.WriteBytes(packet.Data)
	if packet.Client == Shockwave {
		if packet.Header.Dir == Out {
			pkt.WriteInt(2)
		} else {
			pkt.WriteInt(1)
		}
	}
	return pkt
}

func (ext *Ext) mustResolveIdentifier(id Identifier) Header {
	switch id.Dir {
	case In, Out:
	default:
		panic(errors.New("direction must be In or Out"))
	}
	header, ok := ext.headers.TryGet(id)
	if !ok {
		panic(fmt.Errorf("failed to resolve %s header: %q", id.Dir, id.Name))
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

	dbgExt.Printf("initialized (connected: %t)", connected)

	ext.initialized.Dispatch(InitArgs{
		Connected: connected,
	})
}

func (ext *Ext) handleInfoRequest() {
	dbgExt.Println("extension info requested")

	res := &Packet{Header: Header{Out, gOutInfo}}
	res.Write(&ext.info)
	ext.sendRaw(res)
}

func (ext *Ext) handleActivated() {
	dbgExt.Println("extension activated")

	ext.activated.Dispatch()
}

func (ext *Ext) handleConnectionStart(p *Packet) {
	args := ConnectArgs{}
	p.Read(&args.Host, &args.Port, &args.Client, &args.Messages)

	dbgExt.Printf("game connected (%s:%d on %s/%s)", args.Host, args.Port, args.Client.Identifier, args.Client.Version)

	for _, msg := range args.Messages {
		var dir Direction
		if msg.Outgoing {
			dir = Out
		} else {
			dir = In
		}
		ext.headers.Add(msg.Name, Header{dir, uint16(msg.Id)})
	}

	ext.isConnected = true
	ext.isPacketInfoAvailable = true
	ext.remoteHost = args.Host
	ext.remotePort = args.Port
	ext.client = args.Client
	ext.flushInterceptors()

	ext.connectionCtx, ext.closeConnectionCtx = context.WithCancel(context.Background())
	args.Context = ext.connectionCtx

	ext.connected.Dispatch(args)
}

func (ext *Ext) handleConnectionEnd() {
	ext.isConnected = false
	ext.clearIntercepts()

	if close := ext.closeConnectionCtx; close != nil {
		close()
	}
	ext.connectionCtx = nil
	ext.closeConnectionCtx = nil

	dbgExt.Println("game disconnected")

	ext.disconnected.Dispatch()
}

func (ext *Ext) clearIntercepts() {
	ext.interceptsLock.Lock()
	defer ext.interceptsLock.Unlock()
	for msg := range ext.intercepts {
		delete(ext.intercepts, msg)
	}
}

func dispatchIntercept(handlers []InterceptHandler, index int, keep *int, global bool, header Header, intercept *Intercept) (err error) {
	handler := handlers[index]

	defer func() {
		if e := recover(); e != nil {
			err = handlerErr(e, intercept.interceptor, header, global)
		}
	}()

	intercept.Packet.Pos = 0
	handler(intercept)
	if intercept.dereg {
		intercept.dereg = false
	} else {
		handlers[*keep] = handler
		*keep++
	}

	return
}

func (ext *Ext) handlePacketIntercept(p *Packet) (err error) {
	// length of intercept arguments
	length := p.ReadInt()

	tabs := make([]int, 0)
	for i := 4; i < min(length+4, len(p.Data)) && len(tabs) < 3; i++ {
		if p.Data[i] == 0x09 {
			tabs = append(tabs, i)
		}
	}
	if len(tabs) != 3 {
		err = fmt.Errorf("invalid packet intercept data (insufficient delimiter bytes)")
		return
	}

	blocked := p.Data[4] == '1'
	seq, err := strconv.Atoi(string(p.Data[tabs[0]+1 : tabs[1]]))
	if err != nil {
		panic("failed to parse packet index")
	}
	outgoing := p.Data[tabs[1]+3] == 'S'
	modified := p.Data[tabs[2]+1] == '1'

	// Packet offset (starting at the header)
	packetOffset := tabs[2] + 6

	var headerValue uint16
	if ext.client.Type == Shockwave {
		packetOffset = tabs[2] + 2
		headerValue = uint16(encoding.B64Decode(p.Data[packetOffset : packetOffset+2]))
	} else {
		headerValue = binary.BigEndian.Uint16(p.Data[packetOffset:])
	}

	tailOffset := 4 + length
	packetData := p.Data[packetOffset+2 : tailOffset]
	tail := make([]byte, len(p.Data)-tailOffset)
	copy(tail, p.Data[tailOffset:])

	dir := In
	if outgoing {
		dir = Out
	}

	intercept := &Intercept{
		interceptor: ext,
		dir:         dir,
		seq:         seq,
		block:       blocked,
		Packet: &Packet{
			Client: ext.client.Type,
			Header: Header{dir, headerValue},
			Data:   packetData,
		},
	}

	preHeader := intercept.Packet.Header
	preLen := len(packetData)
	checksum := uint32(0)
	if !modified {
		checksum = crc32.ChecksumIEEE(intercept.Packet.Data)
	}

	originalHeader := intercept.Packet.Header
	err = ext.dispatchGlobalIntercepts(originalHeader, intercept)
	if err != nil {
		return
	}
	err = ext.dispatchIntercepts(originalHeader, intercept)
	if err != nil {
		return
	}

	if !modified {
		if preHeader != intercept.Packet.Header ||
			preLen != intercept.Packet.Length() ||
			checksum != crc32.ChecksumIEEE(intercept.Packet.Data) {
			modified = true
		}
	}

	pktModified := intercept.Packet

	// Update the original packet with modified values.
	diff := pktModified.Length() - preLen
	newLen := p.Length() + diff
	p.Header.Value = gOutManipulatedPacket
	p.WriteIntAt(0, newLen-4-len(tail))
	p.WriteByteAt(4, zeroOneChr(intercept.block))
	p.WriteByteAt(tabs[2]+1, zeroOneChr(modified))
	if ext.client.Type != Shockwave {
		p.WriteIntAt(tabs[2]+2, 2+pktModified.Length())
		p.WriteShortAt(packetOffset, int16(pktModified.Header.Value))
	} else {
		encoding.B64Encode(p.Data[packetOffset:packetOffset+2], int(pktModified.Header.Value))
	}
	p.WriteBytesAt(packetOffset+2, pktModified.Data)
	p.WriteBytesAt(tailOffset+diff, tail)
	p.Data = p.Data[:newLen]
	ext.sendRaw(p)
	return
}

func (ext *Ext) dispatchGlobalIntercepts(hdr Header, args *Intercept) (err error) {
	ext.globalInterceptLock.Lock()
	defer ext.globalInterceptLock.Unlock()

	keep := 0
	handlers := ext.globalIntercept.handlers
	for i := range ext.globalIntercept.handlers {
		err = dispatchIntercept(handlers, i, &keep, true, Header{}, args)
		if err != nil {
			return
		}
	}
	ext.globalIntercept.handlers = handlers[:keep]

	return
}

func (ext *Ext) dispatchInterceptGroup(hdr Header, intercept *interceptRegistration, args *Intercept) (err error) {
	if intercept.dereg {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = handlerErr(e, ext, hdr, false)
		}
	}()

	args.Packet.Pos = 0
	intercept.handler(args)

	return
}

func (ext *Ext) snapshotIntercepts(header Header) (snapshot []*interceptRegistration, exists bool) {
	ext.interceptsLock.Lock()
	defer ext.interceptsLock.Unlock()

	src, exists := ext.intercepts[header]
	if exists {
		snapshot = make([]*interceptRegistration, len(src))
		copy(snapshot, src)
	}
	return
}

func (ext *Ext) dispatchIntercepts(hdr Header, args *Intercept) (err error) {
	removals := []*interceptRegistration{}

	header := args.Packet.Header
	if intercepts, exist := ext.snapshotIntercepts(header); exist {
		for _, intercept := range intercepts {
			err = ext.dispatchInterceptGroup(hdr, intercept, args)
			if err != nil {
				return
			}
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
	return
}

func (ext *Ext) removeIntercepts(intercepts ...*interceptRegistration) {
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
			header, ok := ext.headers.TryGet(identifier)
			if !ok {
				continue
			}
			intercepts := ext.intercepts[header]
			i := slices.Index(intercepts, intercept)
			if i != -1 {
				ext.intercepts[header] = slices.Delete(intercepts, i, i+1)
			}
		}
	}
}

func handlerErr(e any, interceptor Interceptor, hdr Header, global bool) error {
	handlerType := "global intercept handler"
	if !global {
		if name, ok := interceptor.Headers().names[hdr]; ok {
			handlerType = fmt.Sprintf("%s %s handler", hdr.Dir.String(), name)
		} else {
			handlerType = fmt.Sprintf("%s (%d) handler", hdr.Dir.String(), hdr.Value)
		}
	}

	return fmt.Errorf("error in %s: %s", handlerType, e)
}
