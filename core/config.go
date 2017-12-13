// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package core

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/shadowsocks/overture/core/cache"
	"github.com/shadowsocks/overture/core/hosts"
	"github.com/shadowsocks/overture/core/outbound"
)

type Config struct {
	BindAddress        string `json:"BindAddress"`
	PrimaryDNS         []*outbound.DNSUpstream
	AlternativeDNS     []*outbound.DNSUpstream
	OnlyPrimaryDNS     bool
	RedirectIPv6Record bool
	IPNetworkFile      string
	DomainFile         string
	DomainBase64Decode bool
	HostsFile          string
	MinimumTTL         int
	CacheSize          int
	RejectQtype        []uint16

	AclList       []*regexp.Regexp
	IPNetworkList []*net.IPNet
	Hosts         *hosts.Hosts
	Cache         *cache.Cache
}

// New config with json file and do some other initiate works
func NewConfig(configFile string) *Config {

	config := parseJson(configFile)

	config.getIPNetworkList()
	config.getDomainList()

	if config.MinimumTTL > 0 {
		log.Info("Minimum TTL is " + strconv.Itoa(config.MinimumTTL))
	} else {
		log.Info("Minimum TTL is disabled")
	}

	config.Cache = cache.New(config.CacheSize)
	if config.CacheSize > 0 {
		log.Info("CacheSize is " + strconv.Itoa(config.CacheSize))
	} else {
		log.Info("Cache is disabled")
	}

	h, err := hosts.New(config.HostsFile)
	if err != nil {
		log.Info("Load hosts file failed: ", err)
	} else {
		config.Hosts = h
		log.Info("Load hosts file successful")
	}

	return config
}

func parseJson(path string) *Config {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Open config file failed: ", err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("Read config file failed: ", err)
		os.Exit(1)
	}

	j := new(Config)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatal("Json syntex error: ", err)
		os.Exit(1)
	}

	return j
}

func (c *Config) getDomainList() {

	var dl []*regexp.Regexp

	f, err := ioutil.ReadFile(c.DomainFile)
	if err != nil {
		log.Error("Open Custom domain file failed: ", err)
		return
	}

	re := regexp.MustCompile(`(\^.+\$)`)
	acl := re.FindAllString(string(f), -1)
	for _, item := range acl {
		re, err := regexp.Compile(item)
		if err == nil {
			dl = append(dl, re)
		}
	}

	if len(dl) > 0 {
		log.Info("Load domain file successful")
	} else {
		log.Warn("There is no element in domain file")
	}

	c.AclList = dl
}

func (c *Config) getIPNetworkList() {

	ipnl := make([]*net.IPNet, 0)
	f, err := os.Open(c.IPNetworkFile)
	if err != nil {
		log.Error("Open IP network file failed: ", err)
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		_, ip_net, err := net.ParseCIDR(s.Text())
		if err != nil {
			break
		}
		ipnl = append(ipnl, ip_net)
	}
	if len(ipnl) > 0 {
		log.Info("Load IP network file successful")
	} else {
		log.Warn("There is no element in IP network file")
	}

	c.IPNetworkList = ipnl
}
