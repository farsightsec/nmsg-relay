/*
 * Copyright (c) 2018 Farsight Security, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package main

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/farsightsec/go-nmsg"
	_ "github.com/farsightsec/go-nmsg/nmsg_base"
	_ "github.com/farsightsec/go-nmsg_sie"
	"github.com/farsightsec/sielink/client"
)

func startClient(config *Config) client.Client {
	cconfig := &client.Config{
		Heartbeat: config.Heartbeat.Duration,
		URL:       "http://localhost/nmsg-relay",
		APIKey:    config.APIKey.String(),
	}

	cli := client.NewClient(cconfig)

	for _, s := range config.Servers {
		if !strings.HasPrefix(s.Path, "/session/") {
			s.Path = "/session/nmsg-relay-upload"
		}
		go func(uri string) {
			for {
				log.Printf("connecting to %s", uri)
				err := cli.DialAndHandle(uri)
				log.Printf("%s: connection closed: %v", uri, err)
				if config.Retry.Duration == 0 {
					return
				}
				<-time.After(config.Retry.Duration)
			}
		}(s.String())
	}
	return cli
}

func runInputLoop(config *Config, input nmsg.Input, output nmsg.Output, addr net.Addr, wg *sync.WaitGroup) {
	for {
		p, err := input.Recv()
		if err != nil {
			if nmsg.IsDataError(err) {
				continue
			}
			log.Printf("Input error on %s: %v", addr, err)
			break
		}
		if config.MsgTypes.Pass(p) {
			output.Send(p)
		}
	}
	output.Close()
	wg.Done()
}

func runInputStats(name string, inputs []nmsg.Input, d time.Duration) {
	var old uint64
	if d == 0 {
		return
	}
	for range time.Tick(d) {
		var total, loss uint64
		for _, input := range inputs {
			stats := input.Stats()
			total += stats.InputContainers
			loss += stats.LostContainers
		}
		log.Printf("%s: Lost %d containers (%d / %d total)",
			name, loss-old, loss, total)
		old = loss
	}
}

func publish(config *Config, cli client.Client) {
	var wg sync.WaitGroup
	var inputs []nmsg.Input

	log.Printf("listening on %s", config.Input.String())
	for _, addr := range config.Input.Addrs() {
		l, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatalf("Failed to open input socket %s: %v", addr.String(), err)
		}
		l.SetReadBuffer(16 * 1048576)
		input := nmsg.NewInput(l, nmsg.MaxContainerSize)
		inputs = append(inputs, input)

		output := nmsg.TimedBufferedOutput(
			newPayloadWriter(config.Channel, cli),
			config.Flush.Duration,
		)
		output.SetMaxSize(nmsg.MaxContainerSize, 16*nmsg.MaxContainerSize)

		wg.Add(1)
		go runInputLoop(config, input, output, addr, &wg)
	}

	go runInputStats(config.Input.String(), inputs, config.StatsInterval.Duration)
	wg.Wait()
	cli.Close()
}

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting nmsg-relay version %s", Version)
	cli := startClient(config)

	publish(config, cli)
}
