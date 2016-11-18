// Copyright (c) 2016 Uber Technologies, Inc.
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

package redis

import (
	"errors"
	"fmt"
	"time"

	redis5 "gopkg.in/redis.v5"
)

type redis5Client struct {
	host string
	port int

	client   *redis5.Client
	queueKey string
}

// NewRedis5Client creates a new QueueClient implementation using gopkg.in/redis.v5
func NewRedis5Client(host string, port int, queueKey string) QueueClient {
	return &redis5Client{
		host:     host,
		port:     port,
		queueKey: queueKey,
	}
}

func (c *redis5Client) Start() error {
	c.client = redis5.NewClient(&redis5.Options{
		Addr: fmt.Sprintf("%s:%d", c.host, c.port),
		DB:   0,
	})

	return c.client.Ping().Err()
}

func (c *redis5Client) Stop() error {
	return c.client.Close()
}

func (c *redis5Client) LPush(item []byte) error {
	cmd := c.client.LPush(c.queueKey, item)
	if cmd.Err() != nil {
		return errors.New("could not push item onto queue")
	}
	return nil
}

type redis5Server struct {
	host string
	port int

	client        *redis5.Client
	queueKey      string
	processingKey string
}

// NewRedis5Server creates a new QueueServer implementation using gopkg.in/redis.v5
func NewRedis5Server(host string, port int, queueKey, processingKey string) QueueServer {
	return &redis5Server{
		host:          host,
		port:          port,
		queueKey:      queueKey,
		processingKey: processingKey,
	}
}

func (c *redis5Server) Start() error {
	c.client = redis5.NewClient(&redis5.Options{
		Addr: fmt.Sprintf("%s:%d", c.host, c.port),
		DB:   0,
	})

	return c.client.Ping().Err()
}

func (c *redis5Server) Stop() error {
	return c.client.Close()
}

func (c *redis5Server) BRPopLPush(timeout time.Duration) ([]byte, error) {
	cmd := c.client.BRPopLPush(c.queueKey, c.processingKey, time.Second)

	item, _ := cmd.Bytes()
	// No bytes means that we timed out waiting for something in our queue
	// and we should try again
	if len(item) == 0 {
		return nil, errors.New("no item found in queue")
	}

	return item, nil
}

// LRem removes item from the queue
func (c *redis5Server) LRem(item []byte) error {
	removed := c.client.LRem(c.processingKey, 1, item).Val()
	if removed <= 0 {
		return errors.New("could not remove item from queue")
	}

	return nil
}
