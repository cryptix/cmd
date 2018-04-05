package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"cryptoscope.co/go/luigi"
	"cryptoscope.co/go/muxrpc"
	"cryptoscope.co/go/muxrpc/codec"
	"cryptoscope.co/go/secretstream"
	"github.com/cryptix/go/debug"
	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

func ssbTryGossip(ctx context.Context, addrWithKey string) error {
	start := time.Now()
	args := strings.SplitN(addrWithKey, ":", 3)
	if len(args) != 3 {
		return errors.Errorf("tryGossip: expected host:port:key. got: %v", args)
	}

	c, err := secretstream.NewClient(*localKey, sbotAppKey)
	if err != nil {
		return err
	}
	var rk = args[2]
	var remotPubKey = localKey.Public
	rk = strings.TrimSuffix(rk, ".ed25519")
	rk = strings.TrimPrefix(rk, "@")
	rpk, err := base64.StdEncoding.DecodeString(rk)
	if err != nil {
		return errors.Wrapf(err, "tryGossip: base64 decode of remoteKey failed")
	}
	copy(remotPubKey[:], rpk)
	d, err := c.NewDialer(remotPubKey)
	if err != nil {
		return err
	}
	addr := strings.Join(args[:2], ":")
	conn, err := d("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "tryGossip: dialing %q failed", addr)
	}
	log.Log("try", "gossip", "addr", addr)

	counter := debug.WrapCounter(conn)
	p := muxrpc.NewPacker(counter)
	if verboseLogging {
		p = muxrpc.NewPacker(codec.Wrap(kitlog.With(log, "id", args[2]), counter))
	}
	gossipHandler := sbotHandler{args[2]}
	rpc := muxrpc.Handle(p, gossipHandler)

	go serveRpc(ctx, start, args[2], rpc, counter)

	type msg struct {
		Key   string `json:"key"`
		Value struct {
			Author    string          `json:"author"`
			Signature string          `json:"signature"`
			Timestamp float64         `json:"timestamp"`
			Seq       float64         `json:"sequence"`
			Content   json.RawMessage `json:"content"`
		} `json:"value"`
	}
	type histArg struct {
		Id   string `json:"id"`
		Seq  int    `json:"seq"`
		Keys bool   `json:"keys"`
		Live bool   `json:"live"`
	}

	src, err := rpc.Source(ctx, msg{}, []string{"createHistoryStream"}, histArg{Id: localID, Seq: 0, Keys: true})
	if err != nil {
		return errors.Wrapf(err, "tryGossip: createHistoryStream failed")
	}

	var closed bool
	for !closed {
		v, err := src.Next(ctx)
		if err != nil {
			if luigi.IsEOS(err) {
				closed = true
				continue
			} else {
				return errors.Wrapf(err, "tryGossip: src.Next failed")
			}
		}

		m := v.(msg)
		log.Log("draining", "myHist", "seq", m.Value.Seq, "k", m.Key)
	}

	return nil
}
