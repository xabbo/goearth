package goearth

import (
	"context"
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

/* Intercept args */

type InterceptArgs struct {
	ext    *Ext
	dir    Direction
	seq    int
	dereg  bool
	block  bool
	Packet *Packet // The intercepted packet.
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

// Blocks the intercepted packet.
func (args *InterceptArgs) Block() {
	args.block = true
}

// Deregisters the current intercept handler.
func (args *InterceptArgs) Deregister() {
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
	ext         *Ext
	transient   bool
	identifiers map[Identifier]struct{}
}

func (b interceptBuilder) Transient() InterceptBuilder {
	b.transient = true
	return b
}

func (b interceptBuilder) With(handler InterceptHandler) InterceptRef {
	identifiers := make(map[Identifier]struct{}, len(b.identifiers))
	maps.Copy(identifiers, b.identifiers)

	grp := &interceptGroup{
		ext:         b.ext,
		identifiers: identifiers,
		handler:     handler,
	}

	b.ext.registerInterceptGroup(grp, b.transient, true)

	return grp
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
	ext         *Ext
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

// Handles the intercept logic for an inline interceptor.
func (i *inlineInterceptor) interceptHandler(e *InterceptArgs) {
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
		i.ref = i.ext.Intercept(i.identifiers...).Transient().With(i.interceptHandler)
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

// Represents an intercept handler with a group of identifiers.
type interceptGroup struct {
	ext         *Ext
	identifiers map[Identifier]struct{}
	handler     InterceptHandler
	dereg       bool
}

func (intercept *interceptGroup) Deregister() {
	intercept.ext.removeIntercepts(intercept)
}
