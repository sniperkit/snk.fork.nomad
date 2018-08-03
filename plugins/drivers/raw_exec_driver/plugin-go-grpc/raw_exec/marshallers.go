package raw_exec

import (
	"net"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/nomad/plugins/drivers/raw_exec_driver/proto"
)

// PluginConfig returns a config from an ExecutorReattachConfig
func unMarshallPluginReattachConfig(c *proto.PluginReattachInfo) *plugin.ReattachConfig {
	var addr net.Addr
	switch c.AddressNetwork {
	case "unix", "unixgram", "unixpacket":
		addr, _ = net.ResolveUnixAddr(c.AddressNetwork, c.AddressName)
	case "tcp", "tcp4", "tcp6":
		addr, _ = net.ResolveTCPAddr(c.AddressNetwork, c.AddressName)
	}
	return &plugin.ReattachConfig{Pid: int(c.Pid), Addr: addr}
}

func marshallPluginReattachConfig(c *plugin.ReattachConfig) *proto.PluginReattachInfo {
	return &proto.PluginReattachInfo{Pid: uint32(c.Pid), AddressNetwork: string(c.Addr.Network()), AddressName: c.Addr.String()}
}
