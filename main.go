// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This is Platina's Mk1 TOR.
//
// To build this source you'll first need to extract the driver plugin from an
// existing program binary.
//
//	unzip goes-platina-mk1 fe1.so
//	zip fe1.zip fe1.so
//
// Or with NDA access to the plugin source, build it with,
//
//	go build -buildmode=plugin github.com/platinasystems/fe1/fe1
//	zip fe1.zip fe1.so
//
// Then build the program and append the zipped plugin.
//
//	go build
//	cat fe1.zip >> goes-platina-mk1
//	zip -q -A goes-platina-mk1
package main

import (
	"plugin"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/platform/mk1"
	"github.com/platinasystems/vnet"
	fe1 "github.com/platinasystems/vnet/devices/ethernet/switch/fe1"
	platform "github.com/platinasystems/vnet/platforms/fe1"
)

type cache struct {
	plugin   *plugin.Plugin
	licenses map[string]string
	patents  map[string]string
	versions map[string]string
	init     func(*vnet.Vnet, *platform.Platform)
	ap       func(v *vnet.Vnet, pp *platform.Platform)
}

func main() {
	var c cache
	goes.Info.Licenses = func() map[string]string {
		if len(c.licenses) == 0 {
			f := c.symbol("Licenses").(func() map[string]string)
			c.licenses = f()
			c.licenses["goes"] = goes.License
		}
		return c.licenses
	}
	goes.Info.Patents = func() map[string]string {
		if len(c.patents) == 0 {
			f := c.symbol("Patents").(func() map[string]string)
			c.patents = f()
			c.patents["goes"] = goes.Patents
		}
		return c.patents
	}
	goes.Info.Versions = func() map[string]string {
		if len(c.versions) == 0 {
			f := c.symbol("Versions").(func() map[string]string)
			c.versions = f()
			c.versions["goes-platina-mk1"] = Version
		}
		return c.versions
	}
	fe1.Init = func(v *vnet.Vnet, p *platform.Platform) {
		if c.init == nil {
			c.init = c.symbol("Init").(func(*vnet.Vnet,
				*platform.Platform))
		}
		c.init(v, p)
	}
	fe1.AddPlatform = func(v *vnet.Vnet, p *platform.Platform) {
		if c.ap == nil {
			c.ap = c.symbol("AddPlatform").(func(*vnet.Vnet,
				*platform.Platform))
		}
		c.ap(v, p)
	}
	mk1.Main()
}

func (c *cache) symbol(name string) plugin.Symbol {
	// FIXME first try unpacking zip file appended to Args[0]
	if c.plugin == nil {
		var err error
		c.plugin, err = plugin.Open("/usr/lib/goes/fe1.so")
		if err != nil {
			c.plugin, err = plugin.Open("fe1.so")
			if err != nil {
				panic(err)
			}

		}
	}
	sym, err := c.plugin.Lookup(name)
	if err != nil {
		panic(err)
	}
	return sym
}
