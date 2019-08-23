/*
 * Copyright (c) 2018 Farsight Security, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"

	"github.com/golang/protobuf/proto"
	nmsg "github.com/farsightsec/go-nmsg"
	"github.com/farsightsec/sielink"
	"github.com/farsightsec/sielink/client"
)

type payloadWriter chan []byte

// A payloadWriter packs up data written to it as an sielink Payload
// of type NmsgContainer with the given channel.
func newPayloadWriter(channel uint32, cli client.Client) payloadWriter {
	wchan := make(chan []byte, 10)
	go func() {
		for b := range wchan {
			var c nmsg.Container
			var buf bytes.Buffer
			if err := c.FromBytes(b); err != nil {
				continue
			}
			c.SetCompression(true)
			c.SetMaxSize(nmsg.MaxContainerSize, 2*nmsg.MaxContainerSize)
			if _, err := c.WriteTo(&buf); err != nil {
				continue
			}

			cli.Send(&sielink.Payload{
				Channel:     proto.Uint32(channel),
				PayloadType: sielink.PayloadType_NmsgContainer.Enum(),
				Data:        buf.Bytes(),
			})
		}
	}()
	return wchan
}

func (c payloadWriter) Write(b []byte) (int, error) {
	c <- b
	return len(b), nil
}
