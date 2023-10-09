package goearth

import (
	"context"
	"strings"
	"sync"
	"time"
)

type InterceptArgs struct {
	ext   *Ext
	dir   Direction
	seq   int
	dereg bool
	// The intercepted packet.
	Packet *Packet
	// Whether to block the intercepted message.
	Block bool
}

// Gets the extension that intercepted this message.
func (args *InterceptArgs) Ext() *Ext {
	return args.ext
}

// Gets the direction of the intercepted message.
func (args *InterceptArgs) Dir() Direction {
	return args.dir
}

// Gets the incremental packet sequence number.
func (args *InterceptArgs) Sequence() int {
	return args.seq
}

// Deregisters the current intercept handler.
func (args *InterceptArgs) Deregister() {
	args.dereg = true
}

type InterceptBuilder interface {
	// Registers the intercept handler for the defined incoming messages.
	In(handler InterceptHandler)
	// Registers the intercept handler for the defined outgoing messages.
	Out(handler InterceptHandler)
}

type interceptBuilder struct {
	Ext      *Ext
	Messages []string
}

func (b *interceptBuilder) In(handler InterceptHandler) {
	for _, name := range b.Messages {
		b.Ext.addInterceptIn(name, handler)
	}
}

func (b *interceptBuilder) Out(handler InterceptHandler) {
	for _, name := range b.Messages {
		b.Ext.addInterceptOut(name, handler)
	}
}

type Intercept interface {
	If(cond *func(*Packet) bool) Intercept
	Block() Intercept
	Timeout(duration time.Duration) Intercept
	TimeoutMs(ms time.Duration) Intercept
	TimeoutSec(sec time.Duration) Intercept
	Await() <-chan *Packet
	Wait() *Packet
	Cancel()
}

type intercept struct {
	ext      *Ext
	messages []string
	ctx      context.Context
	cancel   context.CancelFunc
	timeout  *time.Timer
	bind     sync.Once
	block    bool
	cond     *func(*Packet) bool
	result   chan *Packet
}

func (i *intercept) interceptHandler(e *InterceptArgs) {
	select {
	case <-i.ctx.Done():
		// Timed out.
		e.dereg = true
		return
	default:
	}

	if cond := i.cond; cond != nil && !(*cond)(e.Packet) {
		return
	}

	e.dereg = true
	select {
	case <-i.ctx.Done():
		// Timed out.
	case i.result <- e.Packet.Copy():
		if i.block {
			e.Block = true
		}
		i.cancel()
	}
}

func (i *intercept) bindIntercept() {
	i.bind.Do(func() {
		for _, message := range i.messages {
			dir := INCOMING
			if strings.HasPrefix(message, "out:") {
				dir = OUTGOING
				message = message[4:]
			}
			i.ext.addIntercept(dir, message, i.interceptHandler)
		}
	})
}

func (e *Ext) Recv(messages ...string) Intercept {
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan *Packet, 1)
	context.AfterFunc(ctx, func() { close(result) })
	intercept := &intercept{
		ext:      e,
		messages: messages,
		result:   result,
		timeout:  time.AfterFunc(time.Minute, cancel),
		ctx:      ctx,
		cancel:   cancel,
	}
	return intercept
}

func (i *intercept) If(cond *func(*Packet) bool) Intercept {
	i.cond = cond
	return i
}

func (i *intercept) Block() Intercept {
	i.block = true
	return i
}

func (i *intercept) Timeout(duration time.Duration) Intercept {
	if i.timeout.Stop() {
		i.timeout.Reset(duration)
	}
	return i
}

func (i *intercept) TimeoutMs(ms time.Duration) Intercept {
	return i.Timeout(ms * time.Millisecond)
}

func (i *intercept) TimeoutSec(sec time.Duration) Intercept {
	return i.Timeout(sec * time.Second)
}

func (i *intercept) Await() <-chan *Packet {
	i.bindIntercept()
	return i.result
}

func (i *intercept) Wait() *Packet {
	i.bindIntercept()
	return <-i.result
}

func (i *intercept) Cancel() {
	i.cancel()
}
