// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package tchannel

import (
	"time"

	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/peer/hostport"

	"golang.org/x/net/context"
)

type tchannelPeer struct {
	*hostport.Peer
	t        *Transport
	addr     string
	changed  chan struct{}
	released chan struct{}
}

func newPeer(pid hostport.PeerIdentifier, t *Transport) *tchannelPeer {
	return &tchannelPeer{
		addr:     pid.Identifier(),
		Peer:     hostport.NewPeer(pid, t),
		t:        t,
		changed:  make(chan struct{}, 1),
		released: make(chan struct{}, 0),
	}
}

func (p *tchannelPeer) maintainConnection() {
	cancel := func() {}
	defer func() {
		cancel()
	}()

	delay := time.Second

	// Wait for start (so we can be certain that we have a channel).
	<-p.t.once.Started()

	// Attempt to retain an open connection to each peer so long as it is
	// retained.
	for {
		ch := p.t.ch
		pl := ch.RootPeers()
		tp := pl.GetOrAdd(p.addr)

		inbound, outbound := tp.NumConnections()
		if inbound+outbound > 0 {
			p.Peer.SetStatus(peer.Available)
			if !p.waitForChange() {
				break
			}

		} else {
			p.Peer.SetStatus(peer.Connecting)

			// Attempt to connect
			ctx := context.Background()
			ctx, cancel = context.WithTimeout(ctx, p.t.connectionTimeout)
			_, err := tp.Connect(ctx)

			if err == nil {
				p.Peer.SetStatus(peer.Available)
				// Reset on success
				delay = p.t.initialConnectionRetryDelay
				if !p.waitForChange() {
					break
				}
			} else {
				p.Peer.SetStatus(peer.Unavailable)
				// Back-off on fail
				delay *= time.Duration(p.t.connectionRetryBackoffFactor)
				if !p.sleep(delay) {
					break
				}
			}

		}
	}
}

func (p *tchannelPeer) onStatusChanged() {
	select {
	case p.changed <- struct{}{}:
	}
}

func (p *tchannelPeer) waitForChange() bool {
	// Wait for a connection status change
	select {
	case <-p.changed:
		return true
	case <-p.released:
		return false
	case <-p.t.once.Stopped():
		return false
	}
}

func (p *tchannelPeer) sleep(delay time.Duration) bool {
	select {
	case <-time.After(delay):
		return true
	case <-p.released:
		return false
	case <-p.t.once.Stopped():
		return false
	}
}
