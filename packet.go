package goearth

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

var tParsable = reflect.TypeOf((*Parsable)(nil)).Elem()

type Packet struct {
	Header *NamedHeader
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
	*id = pkt.ReadIdPtr(pos)
}

func (length *Length) Parse(pkt *Packet, pos *int) {
	*length = pkt.ReadLengthPtr(pos)
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

func NewPacket(hdr *NamedHeader) *Packet {
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

func (p *Packet) ReadBufferPtr(pos *int, buf []byte) {
	length := len(buf)
	p.assertCanRead(*pos, length)
	copy(buf, p.Data[*pos:])
	*pos += length
}

func (p *Packet) ReadBufferAt(pos int, buf []byte) {
	p.ReadBufferPtr(&pos, buf)
}

func (p *Packet) ReadBuffer(buf []byte) {
	p.ReadBufferPtr(&p.Pos, buf)
}

// Reads a bool from the specified position and advances the position.
func (p *Packet) ReadBoolPtr(pos *int) (value bool) {
	p.assertCanRead(*pos, 1)
	var i int
	if p.Client == Shockwave {
		if vl64lenEncoded(p.Data[*pos]) != 1 {
			panic(fmt.Errorf("attempt to read boolean when VL64 length > 1"))
		}
		i = vl64decode(p.Data[*pos : *pos+1])
	} else {
		i = int(p.Data[*pos])
	}
	if i != 0 && i != 1 {
		panic(fmt.Errorf("attempt to read boolean from non-boolean value: %d", i))
	}
	value = i == 1
	*pos++
	return
}

// Reads a bool from the specified position.
func (p *Packet) ReadBoolAt(pos int) bool {
	return p.ReadBoolPtr(&pos)
}

// Reads a bool from the current position in the packet.
func (p *Packet) ReadBool() bool {
	return p.ReadBoolPtr(&p.Pos)
}

// Reads a byte from the specified position and advances the position.
func (p *Packet) ReadBytePtr(pos *int) (value byte) {
	p.assertCanRead(*pos, 1)
	value = p.Data[*pos]
	*pos++
	return
}

// Reads a byte from the specified position.
func (p *Packet) ReadByteAt(pos int) byte {
	return p.ReadBytePtr(&pos)
}

// Reads a byte from the current position in the packet.
func (p *Packet) ReadByte() byte {
	return p.ReadBytePtr(&p.Pos)
}

// Copies `n` bytes from the specified position and advances the position.
func (p *Packet) ReadBytesPtr(pos *int, n int) (value []byte) {
	p.assertCanRead(*pos, n)
	value = make([]byte, n)
	*pos += copy(value, p.Data[*pos:])
	return
}

// Copies `n` bytes from the specified position.
func (p *Packet) ReadBytesAt(pos int, length int) []byte {
	return p.ReadBytesPtr(&pos, length)
}

// Copies `n` bytes from the current position in the packet.
func (p *Packet) ReadBytes(length int) []byte {
	return p.ReadBytesPtr(&p.Pos, length)
}

// Reads a short from the specified position and advances the position.
func (p *Packet) ReadShortPtr(pos *int) (value int16) {
	p.assertCanRead(*pos, 2)
	value = int16(binary.BigEndian.Uint16(p.Data[*pos:]))
	*pos += 2
	return
}

// Reads a short from the specified position.
func (p *Packet) ReadShortAt(pos int) int16 {
	return p.ReadShortPtr(&pos)
}

// Reads a short from the current position in the packet.
func (p *Packet) ReadShort() int16 {
	return p.ReadShortPtr(&p.Pos)
}

// Reads an int from the specified position and advances the position.
func (p *Packet) ReadIntPtr(pos *int) (value int) {
	if p.Client == Shockwave {
		return p.ReadVL64Ptr(pos)
	}
	p.assertCanRead(*pos, 4)
	value = int(int32(binary.BigEndian.Uint32(p.Data[*pos:])))
	*pos += 4
	return
}

// Reads an int from the specified position.
func (p *Packet) ReadIntAt(pos int) int {
	return p.ReadIntPtr(&pos)
}

// Reads an int from the current position in the packet.
func (p *Packet) ReadInt() int {
	return p.ReadIntPtr(&p.Pos)
}

// Reads a float from the specified position and advances the position.
func (p *Packet) ReadFloatPtr(pos *int) (value float32) {
	if p.Client == Shockwave {
		panic(fmt.Errorf("%w: attempt to read float on Shockwave session", errors.ErrUnsupported))
	}
	p.assertCanRead(*pos, 4)
	bits := binary.BigEndian.Uint32(p.Data[*pos:])
	value = math.Float32frombits(bits)
	*pos += 4
	return
}

// Reads a float from the specified position.
func (p *Packet) ReadFloatAt(pos int) float32 {
	return p.ReadFloatPtr(&pos)
}

// Reads a float from the current position in the packet.
func (p *Packet) ReadFloat() float32 {
	return p.ReadFloatPtr(&p.Pos)
}

// Reads a long from the specified position and advances the position.
func (p *Packet) ReadLongPtr(pos *int) (value int64) {
	if p.Client == Shockwave {
		panic(fmt.Errorf("%w: attempt to read long on Shockwave session", errors.ErrUnsupported))
	}
	p.assertCanRead(*pos, 8)
	x := binary.BigEndian.Uint64(p.Data[*pos:])
	ptr := unsafe.Pointer(&x)
	value = *(*int64)(ptr)
	*pos += 8
	return
}

// Reads a long from the specified position.
func (p *Packet) ReadLongAt(pos int) int64 {
	return p.ReadLongPtr(&pos)
}

// Reads a long from the current position in the packet.
func (p *Packet) ReadLong() int64 {
	return p.ReadLongPtr(&p.Pos)
}

// Reads a string from the specified position and advances the position.
func (p *Packet) ReadStringPtr(pos *int) (value string) {
	if p.Client == Shockwave && p.Header.Dir == In {
		i := *pos
		for i < len(p.Data) && p.Data[i] != 2 {
			i++
		}
		if i >= len(p.Data) {
			panic(fmt.Errorf("unterminated string"))
		}
		value = string(p.Data[*pos:i])
		*pos = i + 1
	} else {
		p.assertCanRead(*pos, 2)
		var length int
		if p.Client == Shockwave {
			length = int(p.ReadB64At(*pos))
		} else {
			length = int(binary.BigEndian.Uint16(p.Data[*pos:]))
		}
		p.assertCanRead(*pos, 2+length)
		value = string(p.Data[*pos+2 : *pos+2+length])
		*pos += (2 + length)
	}
	return
}

// Reads a string from the specified position.
func (p *Packet) ReadStringAt(pos int) string {
	return p.ReadStringPtr(&pos)
}

// Reads a string from the current position in the packet.
func (p *Packet) ReadString() string {
	return p.ReadStringPtr(&p.Pos)
}

// Reads a Length from the specified position and advances the position.
func (p *Packet) ReadLengthPtr(pos *int) (length Length) {
	switch p.Client {
	case Unity, Shockwave:
		return Length(p.ReadShortPtr(pos))
	default:
		return Length(p.ReadIntPtr(pos))
	}
}

// Reads a Length from the specified position.
func (p *Packet) ReadLengthAt(pos int) Length {
	return p.ReadLengthPtr(&pos)
}

// Reads a Length from the current position in the packet.
func (p *Packet) ReadLength() Length {
	return p.ReadLengthPtr(&p.Pos)
}

// Reads an Id from the specified position and advances the position.
func (p *Packet) ReadIdPtr(pos *int) (id Id) {
	switch p.Client {
	case Unity:
		return Id(p.ReadLongPtr(pos))
	case Flash, Shockwave:
		return Id(p.ReadIntPtr(pos))
	default:
		panic(fmt.Errorf("attempt to read Id on unknown client: %s", p.Client))
	}
}

// Reads an Id from the specified position.
func (p *Packet) ReadIdAt(pos int) Id {
	return p.ReadIdPtr(&pos)
}

// Reads an Id from the current position in the packet.
func (p *Packet) ReadId() Id {
	return p.ReadIdPtr(&p.Pos)
}

// Reads values from the specified position and advances the position.
func (p *Packet) ReadPtr(pos *int, values ...any) {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("packet read failed: %v", r))
		}
	}()
	for _, v := range values {
		if !p.readInterfacePtr(pos, v) {
			p.readReflectPtr(pos, reflect.ValueOf(v))
		}
	}
}

// Reads values from the specified position.
func (p *Packet) ReadAt(pos int, values ...any) {
	p.ReadPtr(&pos, values...)
}

// Reads values from the current position in the packet.
func (p *Packet) Read(values ...any) {
	p.ReadPtr(&p.Pos, values...)
}

func (p *Packet) readReflectPtr(pos *int, v reflect.Value) {
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
		p.readReflectPtr(pos, v.Elem())
	case reflect.Array:
		n := v.Len()
		for i := 0; i < n; i++ {
			p.readReflectPtr(&p.Pos, v.Index(i))
		}
	case reflect.Slice:
		t := v.Type()
		len := int(p.ReadLengthPtr(pos))
		slc := reflect.MakeSlice(t, len, len)
		for i := 0; i < len; i++ {
			p.readReflectPtr(pos, slc.Index(i))
		}
		v.Set(slc)
	case reflect.Struct:
		n := v.NumField()
		for i := 0; i < n; i++ {
			if v := v.Field(i); v.CanSet() {
				p.readReflectPtr(pos, v)
			}
		}
	case reflect.Interface:
		if p.readInterfacePtr(pos, v.Interface()) {
			return
		}
	default:
		if v.CanAddr() {
			v = v.Addr()
		}
		if v.CanInterface() {
			if p.readInterfacePtr(pos, v.Interface()) {
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

func (p *Packet) readInterfacePtr(pos *int, value any) bool {
	switch v := value.(type) {
	case Parsable:
		v.Parse(p, pos)
	case *bool:
		*v = p.ReadBoolPtr(pos)
	case *byte:
		*v = p.ReadBytePtr(pos)
	case *int16:
		*v = p.ReadShortPtr(pos)
	case *uint16:
		*v = uint16(p.ReadShortPtr(pos))
	case *int:
		*v = p.ReadIntPtr(pos)
	case *uint:
		*v = uint(p.ReadIntPtr(pos))
	case *int32:
		*v = int32(p.ReadIntPtr(pos))
	case *uint32:
		*v = uint32(p.ReadIntPtr(pos))
	case *float32:
		*v = p.ReadFloatPtr(pos)
	case *float64:
		*v = float64(p.ReadFloatPtr(pos))
	case *int64:
		*v = p.ReadLongPtr(pos)
	case *uint64:
		*v = uint64(p.ReadLongPtr(pos))
	case *string:
		*v = p.ReadStringPtr(pos)
	case []byte:
		p.ReadBufferPtr(pos, v)
	default:
		return false
	}
	return true
}

/* Writing */

// Writes a bool at the specified position and advances the position.
func (p *Packet) WriteBoolPtr(pos *int, value bool) *Packet {
	p.ensureLength(*pos, 1)
	b := uint8(0)
	if value {
		b = 1
	}
	if p.Client == Shockwave {
		p.WriteVL64At(*pos, int(b))
	} else {
		p.Data[*pos] = b
	}
	*pos++
	return p
}

// Writes a bool at the specified position.
func (p *Packet) WriteBoolAt(pos int, value bool) *Packet {
	return p.WriteBoolPtr(&pos, value)
}

// Writes a bool at the current position in the packet.
func (p *Packet) WriteBool(value bool) *Packet {
	return p.WriteBoolPtr(&p.Pos, value)
}

// Writes a byte at the specified position and advances the position.
func (p *Packet) WriteBytePtr(pos *int, value byte) *Packet {
	p.ensureLength(*pos, 1)
	p.Data[*pos] = value
	*pos++
	return p
}

// Writes a byte at the specified position.
func (p *Packet) WriteByteAt(pos int, value byte) *Packet {
	return p.WriteBytePtr(&pos, value)
}

// Writes a byte at the current position in the packet.
func (p *Packet) WriteByte(value byte) *Packet {
	return p.WriteBytePtr(&p.Pos, value)
}

// Writes a slice of bytes at the specified position and advances the position.
func (p *Packet) WriteBytesPtr(pos *int, value []byte) *Packet {
	length := len(value)
	p.ensureLength(*pos, length)
	copy(p.Data[*pos:], value)
	*pos += length
	return p
}

// Writes a slice of bytes at the specified position.
func (p *Packet) WriteBytesAt(pos int, value []byte) *Packet {
	return p.WriteBytesPtr(&pos, value)
}

// Writes a slice of bytes at the current position in the packet.
func (p *Packet) WriteBytes(value []byte) *Packet {
	return p.WriteBytesPtr(&p.Pos, value)
}

// Writes a short at the specified position.
func (p *Packet) WriteShortPtr(pos *int, value int16) *Packet {
	if p.Client == Shockwave {
		p.WriteB64Ptr(pos, int(value))
		return p
	}
	p.ensureLength(*pos, 2)
	binary.BigEndian.PutUint16(p.Data[*pos:], uint16(value))
	*pos += 2
	return p
}

// Writes a short at the specified position.
func (p *Packet) WriteShortAt(pos int, value int16) *Packet {
	return p.WriteShortPtr(&pos, value)
}

// Writes a short at the current position in the packet.
func (p *Packet) WriteShort(value int16) *Packet {
	return p.WriteShortPtr(&p.Pos, value)
}

// Writes an int at the specified position.
func (p *Packet) WriteIntPtr(pos *int, value int) *Packet {
	p.ensureLength(*pos, 4)
	binary.BigEndian.PutUint32(p.Data[*pos:], uint32(value))
	*pos += 4
	return p
}

// Writes an int at the specified position.
func (p *Packet) WriteIntAt(pos, value int) *Packet {
	return p.WriteIntPtr(&pos, value)
}

// Writes an int at the current position in the packet.
func (p *Packet) WriteInt(value int) *Packet {
	return p.WriteIntPtr(&p.Pos, value)
}

// Writes a float at the specified position.
func (p *Packet) WriteFloatPtr(pos *int, value float32) *Packet {
	if p.Client == Shockwave {
		panic(fmt.Errorf("%w: attempt to write float on Shockwave session", errors.ErrUnsupported))
	}
	p.ensureLength(*pos, 4)
	bits := math.Float32bits(value)
	binary.BigEndian.PutUint32(p.Data[*pos:], bits)
	*pos += 4
	return p
}

// Writes a float at the specified position.
func (p *Packet) WriteFloatAt(pos int, value float32) *Packet {
	return p.WriteFloatPtr(&pos, value)
}

// Writes a float at the current position in the packet.
func (p *Packet) WriteFloat(value float32) *Packet {
	return p.WriteFloatPtr(&p.Pos, value)
}

// Writes a long at the specified position.
func (p *Packet) WriteLongPtr(pos *int, value int64) *Packet {
	if p.Client == Shockwave {
		panic(fmt.Errorf("%w: attempt to write long on Shockwave session", errors.ErrUnsupported))
	}
	p.ensureLength(*pos, 8)
	binary.BigEndian.PutUint64(p.Data[*pos:], uint64(value))
	*pos += 8
	return p
}

// Writes a long at the specified position.
func (p *Packet) WriteLongAt(pos int, value int64) *Packet {
	return p.WriteLongPtr(&pos, value)
}

// Writes a long at the current position in the packet.
func (p *Packet) WriteLong(value int64) *Packet {
	return p.WriteLongPtr(&p.Pos, value)
}

// Writes a string at the specified position.
func (p *Packet) WriteStringPtr(pos *int, value string) *Packet {
	b := []byte(value)
	size := len(b)

	if p.Client == Shockwave {
		if p.Header.Dir == In {
			p.ensureLength(*pos, 1+size)
			p.WriteBytesPtr(pos, b)
			p.WriteBytePtr(pos, 2)
		} else {
			p.ensureLength(*pos, 2+size)
			p.WriteB64Ptr(pos, size)
			p.WriteBytesPtr(pos, b)
		}
		return p
	}

	p.ensureLength(*pos, 2+size)
	if size >= (1 << 16) {
		panic(fmt.Errorf("string length cannot fit into a uint16"))
	}
	if p.Client == Shockwave {
		p.WriteB64At(*pos, size)
	} else {
		binary.BigEndian.PutUint16(p.Data[*pos:], uint16(size))
	}
	copy(p.Data[*pos+2:], b)
	*pos += (2 + size)
	return p
}

// Writes a string at the specified position.
func (p *Packet) WriteStringAt(pos int, value string) *Packet {
	return p.WriteStringPtr(&pos, value)
}

// Writes a string at the current position in the packet.
func (p *Packet) WriteString(value string) *Packet {
	return p.WriteStringPtr(&p.Pos, value)
}

// Writes a Length at the specified position.
// Length is an int on Flash sessions, int16 on Unity sessions.
func (p *Packet) WriteLengthPtr(pos *int, length Length) *Packet {
	switch p.Client {
	case Flash:
		p.WriteIntPtr(pos, int(length))
	case Unity:
		p.WriteShortPtr(pos, int16(length))
	default:
		panic("Cannot write length: unknown client type.")
	}
	return p
}

// Writes a Length at the specified position.
// Length is an int on Flash sessions, int16 on Unity sessions.
func (p *Packet) WriteLengthAt(pos int, length Length) *Packet {
	return p.WriteLengthPtr(&pos, length)
}

// Writes a Length at the current position in the packet.
// Length is an int on Flash sessions, and an int16 on Unity sessions.
func (p *Packet) WriteLength(length Length) *Packet {
	return p.WriteLengthPtr(&p.Pos, length)
}

// Writes an Id at the specified position.
// Id is an int on Flash sessions, and a long on Unity sessions.
func (p *Packet) WriteIdPtr(pos *int, id Id) *Packet {
	switch p.Client {
	case Flash, Shockwave:
		p.WriteIntPtr(pos, int(id))
	case Unity:
		p.WriteLongPtr(pos, int64(id))
	default:
		panic("Cannot write ID: unknown client type.")
	}
	return p
}

// Writes an Id at the specified position.
// Id is an int on Flash sessions, and a long on Unity sessions.
func (p *Packet) WriteIdAt(pos int, id Id) *Packet {
	return p.WriteIdPtr(&pos, id)
}

// Writes an Id at the current position in the packet.
// Id is an int on Flash sessions, and a long on Unity sessions.
func (p *Packet) WriteId(id Id) *Packet {
	return p.WriteIdPtr(&p.Pos, id)
}

// Writes the specified values at the specified position.
func (p *Packet) WritePtr(pos *int, values ...any) *Packet {
	for _, value := range values {
		switch v := value.(type) {
		case []byte:
			p.WriteBytesPtr(pos, v)
		case bool:
			p.WriteBoolPtr(pos, v)
		case int8:
			p.WriteBytePtr(pos, byte(v))
		case uint8:
			p.WriteBytePtr(pos, v)
		case int16:
			p.WriteShortPtr(pos, v)
		case uint16:
			p.WriteShortPtr(pos, int16(v))
		case int:
			p.WriteIntPtr(pos, v)
		case int32:
			p.WriteIntPtr(pos, int(v))
		case uint32:
			p.WriteIntPtr(pos, int(v))
		case float32:
			p.WriteFloatPtr(pos, v)
		case float64:
			p.WriteFloatPtr(pos, float32(v))
		case int64:
			p.WriteLongPtr(pos, v)
		case uint64:
			p.WriteLongPtr(pos, int64(v))
		case string:
			p.WriteStringPtr(pos, v)
		case Id:
			p.WriteIdPtr(pos, v)
		case Length:
			p.WriteLengthPtr(pos, v)
		case FloatStr:
			p.WriteStringPtr(pos, strconv.FormatFloat(float64(v), 'f', -1, 32))
		default:
			r := reflect.ValueOf(v)
			if r.Kind() == reflect.Pointer {
				r = r.Elem()
			}
			switch r.Kind() {
			case reflect.Struct:
				n := r.NumField()
				for i := 0; i < n; i++ {
					p.WritePtr(pos, r.Field(i).Interface())
				}
			default:
				panic(fmt.Errorf("cannot write type %T to packet: (%+v)", v, v))
			}
		}
	}
	return p
}

// Writes the specified values at the specified position.
func (p *Packet) WriteAt(pos int, values ...any) *Packet {
	return p.WritePtr(&pos, values...)
}

// Writes the specified values at the current position in the packet.
func (p *Packet) Write(values ...any) *Packet {
	return p.WritePtr(&p.Pos, values...)
}

/* Replacement */

// Modifies a string at the specified position.
func (p *Packet) ModifyStringPtr(pos *int, transform func(string) string) *Packet {
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
	if p.Client == Shockwave {
		p.WriteB64Ptr(pos, postlen)
	} else {
		p.WriteShortPtr(pos, int16(postlen))
	}
	p.WriteBytesPtr(pos, postbs)
	return p
}

// Modifies a string at the specified position.
func (p *Packet) ModifyStringAt(pos int, transform func(string) string) *Packet {
	return p.ModifyStringPtr(&pos, transform)
}

// Modifies a string at the current position in the packet.
func (p *Packet) ModifyString(transform func(string) string) *Packet {
	return p.ModifyStringPtr(&p.Pos, transform)
}

// Replaces a string at the specified position.
func (p *Packet) ReplaceStringPtr(pos *int, value string) *Packet {
	return p.ModifyStringPtr(pos, func(_ string) string { return value })
}

// Replaces a string at the specified position.
func (p *Packet) ReplaceStringAt(pos int, value string) *Packet {
	return p.ReplaceStringPtr(&pos, value)
}

// Replaces a string at the current position in the packet.
func (p *Packet) ReplaceString(value string) *Packet {
	return p.ReplaceStringPtr(&p.Pos, value)
}

/* Skipping */

// Skips the specified value types from the current position in the packet.
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
			p.readReflectPtr(&p.Pos, reflect.ValueOf(v))
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

// Shockwave encoding

// Reads a 2-byte B64 value from the specified position.
func (p *Packet) ReadB64Ptr(pos *int) int16 {
	p.assertCanRead(*pos, 2)
	value := 64 * uint16(p.Data[*pos]-0x40)
	value += uint16(p.Data[*pos+1] - 0x40)
	*pos += 2
	return int16(value)
}

// Reads a 2-byte B64 value from the specified position.
func (p *Packet) ReadB64At(pos int) int16 {
	return p.ReadB64Ptr(&pos)
}

// Reads a 2-byte B64 value from the current position in the packet.
func (p *Packet) ReadB64() int16 {
	return p.ReadB64Ptr(&p.Pos)
}

// Writes a 2-byte B64 value at the specified position.
func (p *Packet) WriteB64Ptr(pos *int, value int) *Packet {
	p.ensureLength(*pos, 2)
	b64encode(p.Data[*pos:*pos+2], int(value))
	*pos += 2
	return p
}

// Writes a 2-byte B64 value at the specified position.
func (p *Packet) WriteB64At(pos int, value int) *Packet {
	p.WriteB64Ptr(&pos, value)
	return p
}

// Writes a 2-byte B64 value at the current position in the packet.
func (p *Packet) WriteB64(value int) *Packet {
	p.WriteB64Ptr(&p.Pos, value)
	return p
}

// Writes a VL64 at the specified position.
func (p *Packet) WriteVL64Ptr(pos *int, value int) *Packet {
	n := vl64len(value)
	p.ensureLength(*pos, n)
	vl64encode(p.Data[*pos:], value)
	*pos += n
	return p
}

// Writes a VL64 at the specified position.
func (p *Packet) WriteVL64At(pos int, value int) *Packet {
	p.WriteVL64Ptr(&pos, value)
	return p
}

// Writes a VL64 at the current position in the packet.
func (p *Packet) WriteVL64(value int) *Packet {
	p.WriteVL64Ptr(&p.Pos, value)
	return p
}

// Reads a VL64 from the specified position.
func (p *Packet) ReadVL64Ptr(pos *int) int {
	p.assertCanRead(*pos, 1)
	n := vl64lenEncoded(p.Data[*pos])
	p.assertCanRead(*pos, n)
	value := vl64decode(p.Data[*pos : *pos+n])
	*pos += n
	return value
}

// Reads a VL64 from the specified position.
func (p *Packet) ReadVL64At(pos int) int {
	return p.ReadVL64Ptr(&pos)
}

// Reads a VL64 from the current position in the packet.
func (p *Packet) ReadVL64() int {
	return p.ReadVL64Ptr(&p.Pos)
}
