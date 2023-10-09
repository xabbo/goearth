package goearth

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

var tParsable = reflect.TypeOf((*Parsable)(nil)).Elem()

type Packet struct {
	Header *Header
	Data   []byte
	Pos    int
	Client ClientType
}

// Represents a unique identifier.
// Encoded as an int64 on Unity sessions, int32 on Flash sessions.
type Id int64

// Repesents the length of an array.
// Encoded as an int32 on Unity sessions, int16 on Flash sessions.
type Length int32

// Represents a float32 that is encoded as a string.
type FloatStr float32

// Represents an object that can be parsed from a Packet.
type Parsable interface {
	Parse(pkt *Packet, pos *int)
}

// Represents an object that can be written to a Packet.
type Composable interface {
	Compose(pkt *Packet)
}

func (id *Id) Parse(pkt *Packet, pos *int) {
	*id = pkt.ReadIdAtPtr(pos)
}

func (length *Length) Parse(pkt *Packet, pos *int) {
	*length = pkt.ReadLengthAtPtr(pos)
}

func (floatStr *FloatStr) Parse(pkt *Packet, pos *int) {
	value, err := strconv.ParseFloat(pkt.ReadStringPtr(pos), 32)
	if err != nil {
		panic(err)
	}
	*floatStr = FloatStr(value)
}

// Extends the packet length by `n` bytes.
func (p *Packet) extendLength(n int) {
	p.Data = append(p.Data, make([]byte, n)...)
}

// Ensures the packet has enough capacity to store `n` bytes from the specified position.
func (p *Packet) ensureLength(pos int, n int) {
	if pos < 0 {
		panic("position cannot be < 0.")
	}
	if pos > len(p.Data) {
		panic("position cannot be > packet length.")
	}
	if extend := (pos + n) - len(p.Data); extend > 0 {
		p.Data = append(p.Data, make([]byte, extend)...)
	}
}

// Ensures `n` bytes can be read from the specified position.
func (p *Packet) assertCanRead(pos int, n int) {
	if pos < 0 {
		panic(fmt.Errorf("packet position cannot be < 0"))
	}
	if (pos + n) > len(p.Data) {
		panic(fmt.Errorf("attempt to read past packet length"))
	}
}

func NewPacket(hdr *Header) *Packet {
	return &Packet{
		Header: hdr,
		Data:   make([]byte, 0, 8),
	}
}

// Gets the length of the packet's data.
func (p *Packet) Length() int {
	return len(p.Data)
}

/* Reading */

func (p *Packet) ReadBufferAtPtr(pos *int, buf []byte) {
	length := len(buf)
	p.assertCanRead(*pos, length)
	copy(buf, p.Data[*pos:])
	*pos += length
}

func (p *Packet) ReadBufferAt(pos int, buf []byte) {
	p.ReadBufferAtPtr(&pos, buf)
}

func (p *Packet) ReadBuffer(buf []byte) {
	p.ReadBufferAtPtr(&p.Pos, buf)
}

// Reads a bool from the specified position in the packet and advances the position.
func (p *Packet) ReadBoolAtPtr(pos *int) (value bool) {
	p.assertCanRead(*pos, 1)
	value = p.Data[*pos] != 0
	*pos++
	return
}

// Reads a bool from the specified position in the packet.
func (p *Packet) ReadBoolAt(pos int) bool {
	return p.ReadBoolAtPtr(&pos)
}

// Reads a bool from the current position in the packet.
func (p *Packet) ReadBool() bool {
	return p.ReadBoolAtPtr(&p.Pos)
}

// Reads a byte from the specified position in the packet and advances the position.
func (p *Packet) ReadByteAtPtr(pos *int) (value byte) {
	p.assertCanRead(*pos, 1)
	value = p.Data[*pos]
	*pos++
	return
}

// Reads a byte from the specified position in the packet.
func (p *Packet) ReadByteAt(pos int) byte {
	return p.ReadByteAtPtr(&pos)
}

// Reads a byte from the current position in the packet.
func (p *Packet) ReadByte() byte {
	return p.ReadByteAtPtr(&p.Pos)
}

// Copies `n` bytes from the specified position in the packet and advances the position.
func (p *Packet) ReadBytesAtPtr(pos *int, n int) (value []byte) {
	p.assertCanRead(*pos, n)
	value = make([]byte, n)
	*pos += copy(value, p.Data[*pos:])
	return
}

// Copies `n` bytes from the specified position in the packet.
func (p *Packet) ReadBytesAt(pos int, length int) []byte {
	return p.ReadBytesAtPtr(&pos, length)
}

// Copies `n` bytes from the current position in the packet.
func (p *Packet) ReadBytes(length int) []byte {
	return p.ReadBytesAtPtr(&p.Pos, length)
}

// Reads a short from the specified position in the packet and advances the position.
func (p *Packet) ReadShortAtPtr(pos *int) (value int16) {
	p.assertCanRead(*pos, 2)
	value = int16(binary.BigEndian.Uint16(p.Data[*pos:]))
	*pos += 2
	return
}

// Reads a short from the specified position in the packet.
func (p *Packet) ReadShortAt(pos int) int16 {
	return p.ReadShortAtPtr(&pos)
}

// Reads a short from the current position in the packet.
func (p *Packet) ReadShort() int16 {
	return p.ReadShortAtPtr(&p.Pos)
}

// Reads an int from the specified position in the packet and advances the position.
func (p *Packet) ReadIntPtr(pos *int) (value int) {
	p.assertCanRead(*pos, 4)
	value = int(int32(binary.BigEndian.Uint32(p.Data[*pos:])))
	*pos += 4
	return
}

// Reads an int from the specified position in the packet.
func (p *Packet) ReadIntAt(pos int) int {
	return p.ReadIntPtr(&pos)
}

// Reads an int from the current position in the packet.
func (p *Packet) ReadInt() int {
	return p.ReadIntPtr(&p.Pos)
}

// Reads a float from the specified position in the packet and advances the position.
func (p *Packet) ReadFloatAtPtr(pos *int) (value float32) {
	p.assertCanRead(*pos, 4)
	bits := binary.BigEndian.Uint32(p.Data[*pos:])
	value = math.Float32frombits(bits)
	*pos += 4
	return
}

// Reads a float from the specified position in the packet.
func (p *Packet) ReadFloatAt(pos int) float32 {
	return p.ReadFloatAtPtr(&pos)
}

// Reads a float from the current position in the packet.
func (p *Packet) ReadFloat() float32 {
	return p.ReadFloatAtPtr(&p.Pos)
}

// Reads a long from the specified position in the packet and advances the position.
func (p *Packet) ReadLongAtPtr(pos *int) (value int64) {
	p.assertCanRead(*pos, 8)
	x := binary.BigEndian.Uint64(p.Data[*pos:])
	ptr := unsafe.Pointer(&x)
	value = *(*int64)(ptr)
	*pos += 8
	return
}

// Reads a long from the specified position in the packet.
func (p *Packet) ReadLongAt(pos int) int64 {
	return p.ReadLongAtPtr(&pos)
}

// Reads a long from the current position in the packet.
func (p *Packet) ReadLong() int64 {
	return p.ReadLongAtPtr(&p.Pos)
}

// Reads a string from the specified position in the packet and advances the position.
func (p *Packet) ReadStringPtr(pos *int) (value string) {
	p.assertCanRead(*pos, 2)
	length := int(binary.BigEndian.Uint16(p.Data[*pos:]))
	p.assertCanRead(*pos, 2+length)
	value = string(p.Data[*pos+2 : *pos+2+length])
	*pos += (2 + length)
	return
}

// Reads a string from the specified position in the packet.
func (p *Packet) ReadStringAt(pos int) string {
	return p.ReadStringPtr(&pos)
}

// Reads a string from the current position in the packet.
func (p *Packet) ReadString() string {
	return p.ReadStringPtr(&p.Pos)
}

// Reads a Length from the specified position in the packet and advances the position.
func (p *Packet) ReadLengthAtPtr(pos *int) (length Length) {
	if p.Client == UNITY {
		return Length(p.ReadShortAtPtr(pos))
	} else {
		return Length(p.ReadIntPtr(pos))
	}
}

// Reads a Length from the specified position in the packet.
func (p *Packet) ReadLengthAt(pos int) Length {
	return p.ReadLengthAtPtr(&pos)
}

// Reads a Length from the current position in the packet.
func (p *Packet) ReadLength() Length {
	return p.ReadLengthAtPtr(&p.Pos)
}

// Reads an Id from the specified position in the packet and advances the position.
func (p *Packet) ReadIdAtPtr(pos *int) (id Id) {
	if p.Client == UNITY {
		return Id(p.ReadLongAtPtr(pos))
	} else {
		return Id(p.ReadIntPtr(pos))
	}
}

// Reads an Id from the specified position in the packet.
func (p *Packet) ReadIdAt(pos int) Id {
	return p.ReadIdAtPtr(&pos)
}

// Reads an Id from the current position in the packet.
func (p *Packet) ReadId() Id {
	return p.ReadIdAtPtr(&p.Pos)
}

// Reads values from the specified position in the packet and advances the position.
func (p *Packet) ReadPtr(pos *int, values ...any) {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("packet read failed: %v", r))
		}
	}()
	for _, v := range values {
		if !p.readInterfaceAt(pos, v) {
			p.readReflectAt(pos, reflect.ValueOf(v))
		}
	}
}

// Reads values from the specified position in the packet.
func (p *Packet) ReadAt(pos int, values ...any) {
	p.ReadPtr(&pos, values...)
}

// Reads values from the current position in the packet.
func (p *Packet) Read(values ...any) {
	p.ReadPtr(&p.Pos, values...)
}

func (p *Packet) readReflectAt(pos *int, v reflect.Value) {
	parsable := false
	if v.CanAddr() {
		a := v.Addr()
		parsable = a.Type().Implements(tParsable)
	}
	// fmt.Printf("readReflect(%d, %s) parsable=%t\n", *pos, v.Type().String(), parsable)
	if parsable {
		a := v.Addr().Interface().(Parsable)
		a.Parse(p, pos)
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		p.readReflectAt(pos, v.Elem())
	case reflect.Array:
		n := v.Len()
		for i := 0; i < n; i++ {
			p.readReflectAt(&p.Pos, v.Index(i))
		}
	case reflect.Slice:
		t := v.Type()
		len := int(p.ReadLengthAtPtr(pos))
		slc := reflect.MakeSlice(t, len, len)
		for i := 0; i < len; i++ {
			p.readReflectAt(pos, slc.Index(i))
		}
		v.Set(slc)
	case reflect.Struct:
		n := v.NumField()
		for i := 0; i < n; i++ {
			if v := v.Field(i); v.CanSet() {
				p.readReflectAt(pos, v)
			}
		}
	case reflect.Interface:
		if p.readInterfaceAt(pos, v.Interface()) {
			return
		}
	default:
		if v.CanAddr() {
			v = v.Addr()
		}
		if v.CanInterface() {
			if p.readInterfaceAt(pos, v.Interface()) {
				return
			}
		}
		/*if v.CanAddr() {
			if p.readInterfaceAt(pos, v.Addr().Interface()) {
				return
			}
		}*/
		panic(fmt.Errorf("cannot read unsupported type: %+v", v.Type()))
	}
}

func (p *Packet) readInterfaceAt(pos *int, value any) bool {
	switch v := value.(type) {
	case Parsable:
		v.Parse(p, pos)
	case *bool:
		*v = p.ReadBoolAtPtr(pos)
	case *byte:
		*v = p.ReadByteAtPtr(pos)
	case *int16:
		*v = p.ReadShortAtPtr(pos)
	case *uint16:
		*v = uint16(p.ReadShortAtPtr(pos))
	case *int:
		*v = p.ReadIntPtr(pos)
	case *uint:
		*v = uint(p.ReadIntPtr(pos))
	case *int32:
		*v = int32(p.ReadIntPtr(pos))
	case *uint32:
		*v = uint32(p.ReadIntPtr(pos))
	case *float32:
		*v = p.ReadFloatAtPtr(pos)
	case *float64:
		*v = float64(p.ReadFloatAtPtr(pos))
	case *int64:
		*v = p.ReadLongAtPtr(pos)
	case *uint64:
		*v = uint64(p.ReadLongAtPtr(pos))
	case *string:
		*v = p.ReadStringPtr(pos)
	case []byte:
		p.ReadBufferAtPtr(pos, v)
	default:
		return false
	}
	return true
}

/* Writing */

// Writes a bool at the specified position in the packet and advances the position.
func (p *Packet) WriteBoolAtPtr(value bool, pos *int) *Packet {
	p.ensureLength(*pos, 1)
	var b uint8 = 0
	if value {
		b = 1
	}
	p.Data[*pos] = b
	*pos++
	return p
}

// Writes a bool at the specified position in the packet.
func (p *Packet) WriteBoolAt(value bool, pos int) *Packet {
	return p.WriteBoolAtPtr(value, &pos)
}

// Writes a bool at the current position in the packet.
func (p *Packet) WriteBool(value bool) *Packet {
	return p.WriteBoolAtPtr(value, &p.Pos)
}

// Writes a byte at the specified position in the packet and advances the position.
func (p *Packet) WriteByteAtPtr(value byte, pos *int) *Packet {
	p.ensureLength(*pos, 1)
	p.Data[*pos] = value
	*pos++
	return p
}

// Writes a byte at the specified position in the packet.
func (p *Packet) WriteByteAt(value byte, pos int) *Packet {
	return p.WriteByteAtPtr(value, &pos)
}

// Writes a byte at the current position in the packet.
func (p *Packet) WriteByte(value byte) *Packet {
	return p.WriteByteAtPtr(value, &p.Pos)
}

// Writes a slice of bytes at the specified position in the packet and advances the position.
func (p *Packet) WriteBytesAtPtr(value []byte, pos *int) *Packet {
	length := len(value)
	p.ensureLength(*pos, length)
	copy(p.Data[*pos:], value)
	*pos += length
	return p
}

// Writes a slice of bytes at the specified position in the packet.
func (p *Packet) WriteBytesAt(value []byte, pos int) *Packet {
	return p.WriteBytesAtPtr(value, &pos)
}

// Writes a slice of bytes at the current position in the packet.
func (p *Packet) WriteBytes(value []byte) *Packet {
	return p.WriteBytesAtPtr(value, &p.Pos)
}

// Writes a short at the specified position in the packet.
func (p *Packet) WriteShortAtPtr(value int16, pos *int) *Packet {
	p.ensureLength(*pos, 2)
	binary.BigEndian.PutUint16(p.Data[*pos:], uint16(value))
	*pos += 2
	return p
}

// Writes a short at the specified position in the packet.
func (p *Packet) WriteShortAt(value int16, pos int) *Packet {
	return p.WriteShortAtPtr(value, &pos)
}

// Writes a short at the current position in the packet.
func (p *Packet) WriteShort(value int16) *Packet {
	return p.WriteShortAtPtr(value, &p.Pos)
}

// Writes an int at the specified position in the packet.
func (p *Packet) WriteIntAtPtr(value int, pos *int) *Packet {
	p.ensureLength(*pos, 4)
	binary.BigEndian.PutUint32(p.Data[*pos:], uint32(value))
	*pos += 4
	return p
}

// Writes an int at the specified position in the packet.
func (p *Packet) WriteIntAt(value int, pos int) *Packet {
	return p.WriteIntAtPtr(value, &pos)
}

// Writes an int at the current position in the packet.
func (p *Packet) WriteInt(value int) *Packet {
	return p.WriteIntAtPtr(value, &p.Pos)
}

// Writes a float at the specified position in the packet.
func (p *Packet) WriteFloatAt(value float32, pos *int) *Packet {
	p.ensureLength(*pos, 4)
	bits := math.Float32bits(value)
	binary.BigEndian.PutUint32(p.Data[*pos:], bits)
	*pos += 4
	return p
}

// Writes a float at the current position in the packet.
func (p *Packet) WriteFloat(value float32) *Packet {
	return p.WriteFloatAt(value, &p.Pos)
}

// Writes a long at the specified position in the packet.
func (p *Packet) WriteLongAt(value int64, pos *int) *Packet {
	p.ensureLength(*pos, 8)
	binary.BigEndian.PutUint64(p.Data[*pos:], uint64(value))
	*pos += 8
	return p
}

// Writes a long at the current position in the packet.
func (p *Packet) WriteLong(value int64) *Packet {
	return p.WriteLongAt(value, &p.Pos)
}

// Writes a string at the specified position in the packet.
func (p *Packet) WriteStringAt(value string, pos *int) *Packet {
	b := []byte(value)
	size := len(b)
	if size >= (1 << 16) {
		panic("String length cannot fit into a uint16")
	}
	p.ensureLength(*pos, 2+size)
	binary.BigEndian.PutUint16(p.Data[*pos:], uint16(size))
	copy(p.Data[*pos+2:], b)
	*pos += (2 + size)
	return p
}

// Writes a string at the current position in the packet.
func (p *Packet) WriteString(value string) *Packet {
	return p.WriteStringAt(value, &p.Pos)
}

// Writes a Length at the specified position in the packet.
// Length is an int on Flash sessions, int16 on Unity sessions.
func (p *Packet) WriteLengthAt(length Length, pos *int) *Packet {
	switch p.Client {
	case FLASH:
		p.WriteIntAtPtr(int(length), pos)
	case UNITY:
		p.WriteShortAtPtr(int16(length), pos)
	default:
		panic("Cannot write length: unknown client type.")
	}
	return p
}

func (p *Packet) WriteLength(length Length) *Packet {
	return p.WriteLengthAt(length, &p.Pos)
}

func (p *Packet) WriteIdAt(id Id, pos *int) *Packet {
	switch p.Client {
	case FLASH:
		p.WriteIntAtPtr(int(id), pos)
	case UNITY:
		p.WriteLongAt(int64(id), pos)
	default:
		panic("Cannot write ID: unknown client type.")
	}
	return p
}

func (p *Packet) WriteId(id Id) *Packet {
	return p.WriteIdAt(id, &p.Pos)
}

// Writes the specified values at the specified position.
func (p *Packet) WriteAtPtr(pos *int, values ...any) *Packet {
	for _, value := range values {
		switch v := value.(type) {
		case []byte:
			p.WriteBytesAtPtr(v, pos)
		case bool:
			p.WriteBoolAtPtr(v, pos)
		case int8:
			p.WriteByteAtPtr(byte(v), pos)
		case uint8:
			p.WriteByteAtPtr(v, pos)
		case int16:
			p.WriteShortAtPtr(v, pos)
		case uint16:
			p.WriteShortAtPtr(int16(v), pos)
		case int:
			p.WriteIntAtPtr(v, pos)
		case int32:
			p.WriteIntAtPtr(int(v), pos)
		case uint32:
			p.WriteIntAtPtr(int(v), pos)
		case float32:
			p.WriteFloatAt(v, pos)
		case float64:
			p.WriteFloatAt(float32(v), pos)
		case int64:
			p.WriteLongAt(v, pos)
		case uint64:
			p.WriteLongAt(int64(v), pos)
		case string:
			p.WriteStringAt(v, pos)
		case Id:
			p.WriteIdAt(v, pos)
		case Length:
			p.WriteLengthAt(v, pos)
		case FloatStr:
			p.WriteStringAt(strconv.FormatFloat(float64(v), 'f', -1, 32), pos)
		default:
			r := reflect.ValueOf(v)
			if r.Kind() == reflect.Pointer {
				r = r.Elem()
			}
			switch r.Kind() {
			case reflect.Struct:
				n := r.NumField()
				for i := 0; i < n; i++ {
					f := r.Field(i)
					switch f.Kind() {
					case reflect.Bool:
						p.WriteBoolAtPtr(f.Bool(), pos)
					case reflect.String:
						p.WriteStringAt(f.String(), pos)
					default:
						panic(fmt.Sprintf("Cannot write struct field of type %v", f.Type()))
					}
				}
			default:
				panic(fmt.Errorf("cannot write type %T to packet: (%+v)", v, v))
			}
		}
	}
	return p
}

func (p *Packet) WriteAt(pos int, values ...any) *Packet {
	return p.WriteAtPtr(&pos, values...)
}

// Writes the specified values at the current position in the packet.
func (p *Packet) Write(values ...any) *Packet {
	return p.WriteAtPtr(&p.Pos, values...)
}

/* Replacement */

func (p *Packet) ModifyStringAtPtr(transform func(string) string, pos *int) *Packet {
	start := *pos
	pre := p.ReadStringPtr(pos)
	end := *pos
	prelen := (end - start - 2)
	post := transform(pre)
	postbs := []byte(post)
	postlen := len(postbs)
	diff := postlen - prelen
	if diff > 0 {
		p.extendLength(diff)
	}
	copy(p.Data[start+2+postlen:], p.Data[end:])
	if diff < 0 {
		p.Data = p.Data[:len(p.Data)+diff]
	}
	*pos = start
	p.WriteShortAtPtr(int16(postlen), pos)
	p.WriteBytesAtPtr(postbs, pos)
	return p
}

func (p *Packet) ModifyStringAt(transform func(string) string, pos int) *Packet {
	return p.ModifyStringAtPtr(transform, &pos)
}

func (p *Packet) ModifyString(transform func(string) string) *Packet {
	return p.ModifyStringAtPtr(transform, &p.Pos)
}

func (p *Packet) ReplaceStringAtPtr(value string, pos *int) *Packet {
	return p.ModifyStringAtPtr(func(_ string) string { return value }, pos)
}

func (p *Packet) ReplaceStringAt(value string, pos int) *Packet {
	return p.ReplaceStringAtPtr(value, &pos)
}

func (p *Packet) ReplaceString(value string) *Packet {
	return p.ReplaceStringAtPtr(value, &p.Pos)
}

/* Skipping */

// Skip 2 ints, a string, then an int:
// p.Skip(0, 0, "", 0)

func (p *Packet) Skip(types ...any) {
	for _, v := range types {
		switch v := v.(type) {
		case Parsable:
			v.Parse(p, &p.Pos)
		case bool:
			p.ReadBool()
		case int8, uint8:
			p.ReadByte()
		case int16, uint16:
			p.ReadShort()
		case int, int32, uint32:
			p.ReadInt()
		case int64, uint64:
			p.ReadLong()
		case float32:
			p.ReadFloat()
		case string:
			p.ReadString()
		default:
			p.readReflectAt(&p.Pos, reflect.ValueOf(v))
		}
	}
}

/* Cloning */

// Makes a copy of the packet.
func (p *Packet) Copy() *Packet {
	data := make([]byte, len(p.Data))
	copy(data, p.Data)
	return &Packet{
		Client: p.Client,
		Header: p.Header,
		Data:   data,
	}
}
