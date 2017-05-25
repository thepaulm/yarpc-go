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

package integrationtest

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/encoding/raw"
	peerbind "go.uber.org/yarpc/peer"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/peer/roundrobin"
	"go.uber.org/yarpc/transport/tchannel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTChannelWithRoundRobin verifies that TChannel appropriately notifies all
// subscribed peer lists when peers become available and unavailable.
// It does so by constructing a round robin peer list backed by the TChannel transport,
// communicating to three servers. One will always work. One will go down
// temporarily. One will be a bogus TCP port that never completes a TChannel
// handshake.
func TestTChannelWithRoundRobin(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	permanent, permanentAddr := server(t, "")
	defer permanent.Stop()

	temporary, temporaryAddr := server(t, "")
	defer temporary.Stop()

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err, "listen for bogus server")
	invalidAddr := l.Addr().String()
	defer l.Close()

	// Construct a client with a bank of peers. We will keep one running all
	// the time. We'll shut one down temporarily. One will be invalid.
	// The round robin peer list should only choose peers that have
	// successfully connected.
	client, c := client(t, []string{
		permanentAddr,
		temporaryAddr,
		invalidAddr,
	})
	defer client.Stop()

	// All requests should succeed. The invalid peer never enters the rotation.
	blast(ctx, t, c)

	// Shut down one task in the peer list.
	temporary.Stop()
	// One of these requests may fail since one of the peers has gone down but
	// the TChannel transport will not know until a request is attempted.
	call(ctx, c)
	call(ctx, c)
	// All subsequent should succeed since the peer should be removed on
	// connection fail.
	blast(ctx, t, c)

	// Restore the server on the temporary port.
	restored, _ := server(t, temporaryAddr)
	defer restored.Stop()
	blast(ctx, t, c)
}

func blast(ctx context.Context, t *testing.T, c raw.Client) {
	assert.NoError(t, call(ctx, c))
	assert.NoError(t, call(ctx, c))
	assert.NoError(t, call(ctx, c))
	assert.NoError(t, call(ctx, c))
	assert.NoError(t, call(ctx, c))
}

func call(ctx context.Context, c raw.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	res, err := c.Call(ctx, "echo", []byte("hello"))
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(res, []byte("hello")) {
		return fmt.Errorf("unexpected response %+v", res)
	}
	return nil
}

func client(t *testing.T, addrs []string) (*yarpc.Dispatcher, raw.Client) {
	// Convert peer addresses into peer identifiers for a peer list.
	ids := make([]peer.Identifier, len(addrs))
	for i, addr := range addrs {
		ids[i] = hostport.Identify(addr)
	}

	x, err := tchannel.NewTransport(
		tchannel.ServiceName("client"),
		tchannel.ConnectionTimeout(50*time.Millisecond),
		tchannel.InitialConnectionRetryDelay(10*time.Millisecond),
		tchannel.ConnectionRetryBackoffFactor(1),
	)
	require.NoError(t, err, "must construct transport")
	pl := roundrobin.New(x)
	pc := peerbind.Bind(pl, peerbind.BindPeers(ids))
	ob := x.NewOutbound(pc)
	d := yarpc.NewDispatcher(yarpc.Config{
		Name: "client",
		Outbounds: yarpc.Outbounds{
			"service": transport.Outbounds{
				ServiceName: "service",
				Unary:       ob,
			},
		},
	})
	require.NoError(t, d.Start(), "start client dispatcher")
	c := raw.New(d.ClientConfig("service"))
	return d, c
}

func server(t *testing.T, addr string) (*yarpc.Dispatcher, string) {
	x, err := tchannel.NewTransport(
		tchannel.ServiceName("service"),
		tchannel.ListenAddr(addr),
	)
	require.NoError(t, err, "must construct transport")
	ib := x.NewInbound()
	d := yarpc.NewDispatcher(yarpc.Config{
		Name:     "service",
		Inbounds: yarpc.Inbounds{ib},
	})

	handle := func(ctx context.Context, req []byte) ([]byte, error) {
		return req, nil
	}

	d.Register(raw.Procedure("echo", handle))
	require.NoError(t, d.Start(), "start server dispatcher")
	return d, x.ListenAddr()
}
