package raw_exec

import (
	"net"

	"fmt"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/nomad/plugins/drivers/raw_exec_driver/proto"
)

// PluginConfig returns a config from an ExecutorReattachConfig
func unMarshallPluginReattachConfig(c *proto.PluginReattachInfo) *plugin.ReattachConfig {
	//TODO(preetha): Remove this hack! For some reason go-plugin's address network
	// from client.ReattachConfig returns netrpc instead of unix
	c.AddressNetwork = "unix"
	fmt.Printf("Before unmarshalling %+v\n", c)
	var addr net.Addr
	switch c.AddressNetwork {
	case "unix", "unixgram", "unixpacket":
		fmt.Println("**** in resolve unix addr ****")
		addr, _ = net.ResolveUnixAddr(c.AddressNetwork, c.AddressName)
	case "tcp", "tcp4", "tcp6":
		addr, _ = net.ResolveTCPAddr(c.AddressNetwork, c.AddressName)
	}
	return &plugin.ReattachConfig{Pid: int(c.Pid), Addr: addr}
}

func marshallPluginReattachConfig(c *plugin.ReattachConfig) *proto.PluginReattachInfo {
	return &proto.PluginReattachInfo{Pid: uint32(c.Pid), AddressNetwork: string(c.Addr.Network()), AddressName: c.Addr.String()}
}
