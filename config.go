/*
 * Copyright (c) 2023 DomainTools LLC
 * Copyright (c) 2018 Farsight Security, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/farsightsec/go-config"
	"github.com/farsightsec/go-config/env"
	nmsg "github.com/farsightsec/go-nmsg"
	"gopkg.in/yaml.v2"
)

type mType struct {
	vid   uint32
	mtype uint32
}

type mTypeFilter []mType

func (m *mTypeFilter) Set(s string) error {
	// handle comma-separated values
	cl := strings.Split(s, ",")
	if len(cl) > 1 {
		for _, t := range cl {
			m.Set(t)
		}
	}

	s = strings.TrimSpace(s)
	l := strings.Split(s, ":")
	if len(l) != 2 {
		return fmt.Errorf("'%s' not in vname:mtype format", s)
	}

	vname, typename := l[0], l[1]
	vid, mtype, err := nmsg.MessageTypeByName(vname, typename)
	if err != nil {
		return err
	}
	*m = append(*m, mType{vid, mtype})
	return nil
}

func (m *mTypeFilter) String() string {
	var l []string
	for _, t := range *m {
		vname, mname, err := nmsg.MessageTypeName(t.vid, t.mtype)
		if err != nil {
			continue
		}
		l = append(l, strings.Join([]string{vname, mname}, ":"))
	}
	return strings.Join(l, ",")
}

func (m *mTypeFilter) Pass(p *nmsg.NmsgPayload) bool {
	if len(*m) == 0 {
		return true
	}
	vid := p.GetVid()
	mtype := p.GetMsgtype()
	for _, f := range *m {
		if vid != f.vid {
			continue
		}
		if mtype != f.mtype {
			continue
		}
		return true
	}
	return false
}

func (m *mTypeFilter) UnmarshalYAML(u func(interface{}) error) error {
	var ss []string
	if err := u(&ss); err != nil {
		return err
	}
	for _, s := range ss {
		if err := m.Set(s); err != nil {
			return err
		}
	}
	return nil
}

// Config represents the global configuration of the client.
type Config struct {
	Servers       []config.URL    `yaml:"servers"`
	APIKey        config.String   `yaml:"api_key"`
	Channel       uint32          `yaml:"channel"`
	Heartbeat     config.Duration `yaml:"heartbeat"`
	Retry         config.Duration `yaml:"retry"`
	Flush         config.Duration `yaml:"flush"`
	Input         nmsg.Sockspec   `yaml:"input"`
	StatsInterval config.Duration `yaml:"stats_interval"`
	MsgTypes      mTypeFilter     `yaml:"message_types"`
}

type uint32val uint32

func (u *uint32val) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 32)
	*u = uint32val(v)
	return err
}
func (u *uint32val) String() string {
	return strconv.FormatUint(uint64(*u), 10)
}

func loadConfig(conf *Config, filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, conf)
}

const envPrefix = "NMSG_RELAY_"

func fixURL(u *url.URL) {
	switch u.Scheme {
	case "wss":
	case "ws":
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "":
		u.Scheme = "wss"
		// A string not starting with "scheme://" is parsed as a relative
		// path. Split this into host, path if needed.
		si := strings.Index(u.Path, "/")
		if si < 0 {
			u.Host = u.Path
			u.Path = ""
			return
		}
		u.Host, u.Path = u.Path[:si], u.Path[si:]
	default:
		// if passed "host:port", url.Parse will treat the hostname as
		// the URL scheme, and the port as an "Opaque" string.
		// Handle that here by prepending "wss://" to the string version
		// of the URL, and re-parsing.
		pu, err := url.Parse("wss://" + u.String())
		if err != nil {
			return
		}
		*u = *pu
	}
}

func parseConfig() (conf *Config, err error) {
	var configFilename string
	var serverList string
	var printVersion bool
	var envConfig = env.NewConfig(true)
	conf = &Config{}

	flag.BoolVar(&printVersion, "v", false, "Print version and exit")
	flag.BoolVar(&printVersion, "version", false, "Print version and exit")
	flag.DurationVar(&conf.Heartbeat.Duration, "heartbeat", 30*time.Second,
		"heartbeat interval (default 30s)")
	flag.DurationVar(&conf.Retry.Duration, "retry", 30*time.Second,
		"connection retry interval (default 30s)")
	flag.DurationVar(&conf.Flush.Duration, "flush", time.Second/2,
		"buffer flush interval (default 500ms)")

	flag.Var(&conf.APIKey, "apikey", "apikey or path to apikey file")
	flag.Var((*uint32val)(&conf.Channel), "channel", "destination channel for NMSG data")
	flag.Var(&conf.Input, "input", "address for datagram input")
	flag.DurationVar(&conf.StatsInterval.Duration, "stats_interval", 0,
		"how often to print input statistics (default 0s / no stats)")
	flag.Var(&conf.MsgTypes, "message_type", "add vname:msgtype to allowed types list (default: allow all types)")

	envConfig.Var(&configFilename, envPrefix+"CONFIG")
	flag.StringVar(&configFilename, "config", configFilename, "read configuration from file")
	flag.Parse()

	if printVersion {
		fmt.Println("This is nmsg-relay, version ", Version)
		os.Exit(0)
	}

	if configFilename != "" {
		err = loadConfig(conf, configFilename)
		if err != nil {
			log.Fatal(err)
		}
	}

	envConfig.Var(&conf.Heartbeat.Duration, envPrefix+"HEARTBEAT")
	envConfig.Var(&conf.Retry.Duration, envPrefix+"RETRY")
	envConfig.Var(&conf.Flush.Duration, envPrefix+"FLUSH")
	envConfig.Var(&conf.APIKey, envPrefix+"APIKEY")
	envConfig.Var((*uint32val)(&conf.Channel), envPrefix+"CHANNEL")
	envConfig.Var(&conf.Input, envPrefix+"INPUT")
	envConfig.Var(&conf.StatsInterval.Duration, envPrefix+"STATS_INTERVAL")
	envConfig.Var(&serverList, envPrefix+"SERVERS")
	envConfig.Var(&conf.MsgTypes, envPrefix+"MESSAGE_TYPES")

	if configFilename != "" {
		// Parse flags again to override configuration and environment
		// values.
		flag.Parse()
	}

	if len(serverList) > 0 {
		for _, s := range strings.Split(serverList, " ") {
			var u config.URL
			if err := u.Set(s); err != nil {
				log.Fatalf("Invalid %s value %s: %v",
					envPrefix+"SERVERS", s, err)
			}
			fixURL(u.URL)
			conf.Servers = append(conf.Servers, u)
		}
	}

	if flag.NArg() > 0 {
		var servers []config.URL
		for _, s := range flag.Args() {
			u := config.URL{}
			if perr := u.Set(s); perr != nil {
				err = fmt.Errorf("Invalid URI %s: %v", s, perr)
				return
			}
			fixURL(u.URL)
			servers = append(servers, u)
		}
		conf.Servers = servers
	}

	if conf.Channel == 0 {
		err = errors.New("no channel specified")
	}
	if len(conf.Servers) == 0 {
		err = errors.New("no servers specified")
	}
	if conf.Input.Addr == nil {
		err = errors.New("no input address specified")
	}
	if conf.APIKey.String() == "" {
		err = errors.New("no API key specified")
	}
	return
}
