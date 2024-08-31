package goearth

import (
	"context"
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

type Interceptor interface {
	Context() context.Context
	Client() Client
	Headers() *Headers
	Send(id Identifier, values ...any)
	SendPacket(*Packet)
	Recv(identifiers ...Identifier) InlineInterceptor
	Register(*InterceptGroup) InterceptRef
	Intercept(...Identifier) InterceptBuilder

	Initialized(EventHandler[InitArgs])
	Connected(EventHandler[ConnectArgs])
	Disconnected(VoidHandler)
}

// Intercept holds the event arguments for an intercepted packet.
type Intercept struct {
	interceptor Interceptor
	dir         Direction
	seq         int
	dereg       bool
	block       bool
	Packet      *Packet // The intercepted packet.
}

func NewIntercept(interceptor Interceptor, packet *Packet, sequence int, blocked bool) *Intercept {
	return &Intercept{
		interceptor: interceptor,
		dir:         packet.Header.Dir,
		seq:         sequence,
		block:       blocked,
		Packet:      packet,
	}
}

// Deprecated: Use [Intercept].
type InterceptArgs = Intercept

// Gets the interceptor that intercepted this message.
func (args *Intercept) Interceptor() Interceptor {
	return args.interceptor
}

// Gets the direction of the intercepted message.
func (args *Intercept) Dir() Direction {
	return args.dir
}

// Name gets the name of the intercepted packet header.
func (args *Intercept) Name() string {
	return args.interceptor.Headers().Name(args.Packet.Header)
}

// Is returns whether the intercepted packet header matches the specified identifier.
func (args *Intercept) Is(id Identifier) bool {
	return args.interceptor.Headers().Is(args.Packet.Header, id)
}

// Gets the incremental packet sequence number.
func (args *Intercept) Sequence() int {
	return args.seq
}

// Blocks the intercepted packet.
func (args *Intercept) Block() {
	args.block = true
}

// IsBlocked gets whether the packet has been flagged to be blocked by the interceptor.
func (args *Intercept) IsBlocked() bool {
	return args.block
}

// Deregisters the current intercept handler.
func (args *Intercept) Deregister() {
	args.dereg = true
}

/* Intercept builder */

type InterceptBuilder interface {
	// Flags the intercept as transient.
	Transient() InterceptBuilder
	// Registers the intercept handler and returns a reference.
	With(handler InterceptHandler) InterceptRef
}

type interceptBuilder struct {
	ix          Interceptor
	transient   bool
	identifiers map[Identifier]struct{}
}

func NewInterceptBuilder(interceptor Interceptor, ids ...Identifier) InterceptBuilder {
	builder := &interceptBuilder{
		ix:          interceptor,
		identifiers: map[Identifier]struct{}{},
	}
	for _, id := range ids {
		builder.identifiers[id] = struct{}{}
	}
	return builder
}

func (b interceptBuilder) Transient() InterceptBuilder {
	b.transient = true
	return b
}

func (b interceptBuilder) With(handler InterceptHandler) InterceptRef {
	identifiers := make(map[Identifier]struct{}, len(b.identifiers))
	maps.Copy(identifiers, b.identifiers)

	grp := &InterceptGroup{
		Identifiers: b.identifiers,
		Handler:     handler,
		Transient:   b.transient,
	}

	return b.ix.Register(grp)
}

/* Inline interceptor */

type InlineInterceptor interface {
	// Configures the intercept condition.
	If(condition func(p *Packet) bool) InlineInterceptor
	// Configures the interceptor to block the intercepted packet.
	Block() InlineInterceptor
	// Configures the timeout duration of the interceptor.
	Timeout(duration time.Duration) InlineInterceptor
	// Configures the timeout duration of the interceptor.
	TimeoutMs(ms time.Duration) InlineInterceptor
	// Configures the timeout duration of the interceptor.
	TimeoutSec(sec time.Duration) InlineInterceptor
	// Returns a channel that will signal the intercepted packet.
	// Returns nil if the interceptor times out or is canceled.
	Await() <-chan *Packet
	// Waits for the packet to be intercepted and then returns it.
	// Returns nil if the interceptor times out or is canceled.
	Wait() *Packet
	// Cancels the interceptor.
	Cancel()
}

type inlineInterceptor struct {
	ix          Interceptor
	identifiers []Identifier
	ctx         context.Context
	cancel      context.CancelFunc
	timeout     *time.Timer
	bindOnce    sync.Once
	block       bool
	cond        func(*Packet) bool
	result      chan *Packet
	ref         InterceptRef
}

func NewInlineInterceptor(interceptor Interceptor, identifiers []Identifier) InlineInterceptor {
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan *Packet, 1)
	context.AfterFunc(ctx, func() { close(result) })
	return &inlineInterceptor{
		ix:          interceptor,
		identifiers: identifiers,
		ctx:         ctx,
		cancel:      cancel,
		timeout:     time.AfterFunc(time.Minute, cancel),
		result:      make(chan *Packet, 1),
	}
}

// Handles the intercept logic for an inline interceptor.
func (i *inlineInterceptor) interceptHandler(e *Intercept) {
	select {
	case <-i.ctx.Done():
		// Timed out.
		e.dereg = true
		return
	default:
	}

	if cond := i.cond; cond != nil && !cond(e.Packet) {
		return
	}

	e.dereg = true
	select {
	case <-i.ctx.Done():
		// Timed out.
	case i.result <- e.Packet.Copy():
		if i.block {
			e.Block()
		}
		i.cancel()
	}
}

func (i *inlineInterceptor) bindIntercept() {
	i.bindOnce.Do(func() {
		i.ref = i.ix.Intercept(i.identifiers...).Transient().With(i.interceptHandler)
	})
}

func (i *inlineInterceptor) If(condition func(p *Packet) bool) InlineInterceptor {
	i.cond = condition
	return i
}

func (i *inlineInterceptor) Block() InlineInterceptor {
	i.block = true
	return i
}

func (i *inlineInterceptor) Timeout(duration time.Duration) InlineInterceptor {
	if i.timeout.Stop() {
		i.timeout.Reset(duration)
	}
	return i
}

func (i *inlineInterceptor) TimeoutMs(ms time.Duration) InlineInterceptor {
	return i.Timeout(ms * time.Millisecond)
}

func (i *inlineInterceptor) TimeoutSec(sec time.Duration) InlineInterceptor {
	return i.Timeout(sec * time.Second)
}

func (i *inlineInterceptor) Await() <-chan *Packet {
	i.bindIntercept()
	return i.result
}

func (i *inlineInterceptor) Wait() *Packet {
	i.bindIntercept()
	return <-i.result
}

func (i *inlineInterceptor) Cancel() {
	i.cancel()
}

/* Intercept reference */

// Represents a reference to an intercept handler.
type InterceptRef interface {
	// Deregisters the intercept handler.
	Deregister()
}

/* Intercept group */

type InterceptGroup struct {
	Identifiers map[Identifier]struct{}
	Handler     InterceptHandler
	Transient   bool
}

// Represents an intercept handler with a group of identifiers.
type interceptRegistration struct {
	ext         *Ext
	identifiers map[Identifier]struct{}
	handler     InterceptHandler
	dereg       bool
}

func (intercept *interceptRegistration) Deregister() {
	intercept.ext.removeIntercepts(intercept)
}
