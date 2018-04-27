package main

import (
	"bytes"
	"encoding/json"
)

type SignedMessage struct {
	Message
	Signature Signature `json:"signature"`
}

type Message struct {
	Previous  *Ref            `json:"previous"`
	Author    Ref             `json:"author"`
	Sequence  int             `json:"sequence"`
	Timestamp float64         `json:"timestamp"`
	Hash      string          `json:"hash"`
	Content   json.RawMessage `json:"content"`
}

func Encode(i interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(i)
	if err != nil {
		return nil, err
	}
	return bytes.Trim(buf.Bytes(), "\n"), nil
}
