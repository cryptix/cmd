package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cryptoscope.co/go/luigi"
	"cryptoscope.co/go/muxrpc"
	"cryptoscope.co/go/ssb"
	"github.com/pkg/errors"
)

type sbotHandler struct {
	remoteID string
}

type retWhoami struct {
	ID string `json:"id"`
}

type createHistArgs struct {
	//map[keys:false id:@Bqm7bG4qvlnWh3BEBFSj2kDr+     30+mUU3hRgrikE2+xc=.ed25519 seq:20 live:true
	Keys bool   `json:"keys"`
	Live bool   `json:"live"`
	Id   string `json:"id"`
	Seq  int    `json:"seq"`
}

func (h sbotHandler) HandleCall(ctx context.Context, req *muxrpc.Request) {
	// TODO: push manifest check into muxrpc
	if req.Type == "" {
		req.Type = "async"
	}

	switch m := strings.Join(req.Method, "."); m {
	case "whoami":
		err := req.Return(ctx, retWhoami{"heinbloed"})
		if err != nil {
			log.Log("call", "whoami", "err", err)
		}

	case "gossip.ping":
		//todo: read args
		go func() {
			for i := 0; i < 3; i++ {
				err := req.Stream.Pour(ctx, time.Now().Unix())
				if err != nil {
					log.Log("call", "gossip.ping", "err", err)
					req.Stream.CloseWithError(errors.Wrap(err, "failed gossiping"))
					return
				}
				log.Log("call", "gossip.ping", "pong", i)
				time.Sleep(1 * time.Second)
			}
			req.Stream.Close()

		}()
		for {
			v, err := req.Stream.Next(ctx)
			if err != nil {
				log.Log("call", "gossip.ping", "err", err)
				req.Stream.CloseWithError(errors.Wrap(err, "failed gossiping"))
				return
			}
			log.Log("call", "gossip.ping", "ping", v)
		}

	case "gossip.connect":
		if len(req.Args) != 1 {
			req.Stream.CloseWithError(errors.Errorf("bad request"))
			return
		}
		addr := req.Args[0].(string)
		ret := make(map[string]interface{})
		ret["addr"] = addr
		err := ssbTryGossip(ctx, addr)
		if err != nil {
			log.Log("try", "gossip.connect", "err", err)
			req.Stream.CloseWithError(errors.Wrap(err, "failed gossiping"))
			return
		} else {
			ret["worked"] = true
		}
		err = req.Return(ctx, ret)
		if err != nil {
			log.Log("call", "gossip.connect", "err", err)
		}

	case "createHistoryStream":
		if len(req.Args) != 1 {
			req.Stream.CloseWithError(errors.Errorf("bad request"))
			return
		}
		_, ok := req.Args[0].(map[string]interface{})
		if !ok {
			log.Log("call", "createHistoryStream", "err", "bad call", "tipe", fmt.Sprintf("%T", req.Args[0]))
			req.Stream.CloseWithError(errors.Errorf("bad args"))
			return
		}
		//var qargs createHistArgs
		//qargs.Keys = qmap["keys"].(bool)
		//qargs.Live = qmap["live"].(bool)
		//qargs.Seq = int(qmap["seq"].(float64))
		//qargs.Id = qmap["id"].(string)
		//fmt.Println("createHist", qargs)
		req.Stream.Close()

	default:
		log.Log("warning", "unhandled call", "method", m, "args", fmt.Sprintf("%+v", req.Args))
		err := errors.Errorf("unhandled call: %s", m)
		// TODO: illegal for async calls to close with Stream
		req.Stream.CloseWithError(err)
	}
}

type RawSignedMessage struct {
	json.RawMessage
}

func (h sbotHandler) HandleConnect(ctx context.Context, e muxrpc.Endpoint) {
	var q = createHistArgs{false, false, h.remoteID, 0}
	source, err := e.Source(ctx, RawSignedMessage{}, []string{"createHistoryStream"}, q)
	if err != nil {
		log.Log("handleConnect", "createHistoryStream", "err", err)
		return
	}
	i := 0
	ref, err := ssb.ParseRef(h.remoteID)
	if err != nil {
		log.Log("handleConnect", "ssb.ParseRef", "err", err)
		return
	}
	for {
		start := time.Now()
		v, err := source.Next(ctx)
		if luigi.IsEOS(err) {
			break
		}
		if err != nil {
			log.Log("handleConnect", "createHistoryStream", "i", i, "err", err)
			break
		}

		rmsg := v.(RawSignedMessage)

		buf, sig, err := ssb.EncodePreserveOrder(rmsg.RawMessage)
		if err != nil {
			err = errors.Wrap(err, "simple Encode failed")
			log.Log("handleConnect", "createHistoryStream", "i", i, "err", err)
			break
		}

		if err := sig.Verify(buf, *ref); err != nil {
			err = errors.Wrap(err, "msg verify failed")
			log.Log("handleConnect", "createHistoryStream", "i", i, "err", err)
			break
		}
		fmt.Printf("\n####\nverified hist%d (took %v):\n%s\n", i, time.Since(start), buf)

		i++
	}
	log.Log("handle", "connect", "Hello", h.remoteID)
}

type retPing struct {
	Pong string
}

func (h sbotHandler) GossipPing(timout int) retPing {
	return retPing{"test"}
}
