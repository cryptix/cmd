package main

import (
	"context"
	"fmt"
	"strings"

	"cryptoscope.co/go/muxrpc"
	"github.com/pkg/errors"
)

type sbotHandler struct {
	remoteID string
}

type retWhoami struct {
	ID string `json:"id"`
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
	default:
		log.Log("warning", "unhandled call", "method", m, "args", fmt.Sprintf("%+v", req.Args))
		req.Stream.CloseWithError(errors.Errorf("unhandled call"))
	}
}

func (h sbotHandler) HandleConnect(ctx context.Context, e muxrpc.Endpoint) {
	/* calling back
	ret, err := e.Async(ctx, "str", []string{"whoami"})
	if err != nil {
		log.Log("handleConnect", "whoami", "err", err)
		return
	}
	*/
	log.Log("handle", "connect", "Hello", h.remoteID)
}

type retPing struct {
	Pong string
}

func (h sbotHandler) GossipPing(timout int) retPing {
	return retPing{"test"}
}
