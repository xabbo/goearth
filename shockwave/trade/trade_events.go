package trade

import g "xabbo.b7c.io/goearth"

// Args holds the arguments for trade events.
type Args struct {
	Opened bool
	Offers Offers
}

// AcceptArgs holds the arguments for trade accept events.
type AcceptArgs struct {
	Name     string
	Accepted bool
}

// Updated registers an event handler that is invoked when the trade is updated.
// If a trade was opened, the Opened flag on the event arguments will be set.
func (mgr *Manager) Updated(handler g.EventHandler[Args]) {
	mgr.updated.Register(handler)
}

// Accepted registers an event handler that is invoked when the trade is accepted.
func (mgr *Manager) Accepted(handler g.EventHandler[AcceptArgs]) {
	mgr.accepted.Register(handler)
}

// Completed registers an event handler that is invoked when the trade is completed.
func (mgr *Manager) Completed(handler g.EventHandler[Args]) {
	mgr.completed.Register(handler)
}

// Closed registers an event that is invoked when the trade is closed.
func (mgr *Manager) Closed(handler g.EventHandler[Args]) {
	mgr.closed.Register(handler)
}
