package profile

import (
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/internal/debug"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
)

var dbg = debug.NewLogger("[profile]")

// Manager tracks the state of the user's profile.
type Manager struct {
	ix               g.Interceptor
	requestOnConnect bool

	updated g.Event[Args]

	Profile
}

func NewManager(ix g.Interceptor) *Manager {
	mgr := &Manager{ix: ix}
	ix.Initialized(mgr.onInitialized)
	ix.Connected(mgr.onConnected)
	ix.Disconnected(mgr.onDisconnected)
	ix.Intercept(in.USER_OBJ).With(mgr.handleUserObj)
	return mgr
}

func (mgr *Manager) onInitialized(e g.InitArgs) {
	if e.Connected {
		// game is already connected, assume safe to send request
		mgr.requestOnConnect = true
		dbg.Println("game is connected, will request on connect")
	}
	// otherwise wait to receive the USER_OBJ packet naturally
}

func (mgr *Manager) onConnected(e g.ConnectArgs) {
	if mgr.requestOnConnect {
		mgr.ix.Send(out.INFORETRIEVE)
		dbg.Println("requested profile")
	}
}

func (mgr *Manager) onDisconnected() {
	mgr.requestOnConnect = false
}

// handlers

func (mgr *Manager) handleUserObj(e *g.Intercept) {
	e.Packet.Read(&mgr.Profile)
	mgr.updated.Dispatch(Args{mgr.Profile})

	dbg.Printf("received user profile for %q", mgr.Profile.Name)
}
