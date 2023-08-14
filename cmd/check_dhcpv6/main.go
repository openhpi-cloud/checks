package main

import (
	"fmt"
	"net"
	"os"
	"time"

	arg "github.com/alexflint/go-arg"
	nagios "github.com/atc0005/go-nagios"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/dhcpv6/client6"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {
	var opts struct {
		Interface string `arg:"-i,--interface,required" help:"Interface to send DHCPv6 solicit"`
		Debug     bool   `arg:"--debug"`
	}

	arg.MustParse(&opts)

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	var plugin = nagios.NewPlugin()
	plugin.ExitStatusCode = nagios.StateUNKNOWNExitCode
	defer plugin.ReturnCheckResults()

	iface, err := net.InterfaceByName(opts.Interface)
	if err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}

	llAddr, err := getLinkLocalAddress(iface)
	if err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}

	laddr := net.UDPAddr{
		IP:   llAddr,
		Port: dhcpv6.DefaultClientPort,
		Zone: opts.Interface,
	}

	raddr := net.UDPAddr{
		IP:   dhcpv6.AllDHCPRelayAgentsAndServers,
		Port: dhcpv6.DefaultServerPort,
	}

	conn, err := net.ListenUDP("udp6", &laddr)
	if err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}
	defer conn.Close()

	// wait for the listener to be ready, fail if it takes too much time
	deadline := time.Now().Add(time.Second)
	for {
		if now := time.Now(); now.After(deadline) {
			return
		}
		if conn.LocalAddr() != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}

	solicit, err := dhcpv6.NewSolicit(iface.HardwareAddr)
	if err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}

	log.Debugf("%s\n", solicit.Summary())

	_, err = conn.WriteTo(solicit.ToBytes(), &raddr)
	if err != nil {
		plugin.AddError(err)
		plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
		return
	}

	var adv dhcpv6.DHCPv6
	buf := make([]byte, client6.MaxUDPReceivedPacketSize)
	oobdata := []byte{}
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		n, _, _, _, err := conn.ReadMsgUDP(buf, oobdata)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				plugin.ServiceOutput = fmt.Sprint("No advertise message received")
				plugin.ExitStatusCode = nagios.StateCRITICALExitCode
				return
			}

			plugin.AddError(err)
			plugin.ServiceOutput = fmt.Sprintf("ERR - %s", err)
			return
		}

		adv, err = dhcpv6.FromBytes(buf[:n])
		if err != nil {
			// skip non-DHCP packets
			//
			// TODO: It also skips DHCP packets with any errors (for
			// example if bootfile params are encoded incorrectly). We
			// need to log such cases instead of silently skip them.
			continue
		}

		if recvMsg, ok := adv.(*dhcpv6.Message); ok {
			// Check transaction ID if reply to send solicit message
			if solicit.TransactionID != recvMsg.TransactionID {
				continue
			}
		}

		if adv.Type() == dhcpv6.MessageTypeAdvertise {
			break
		}
	}

	log.Debugf(adv.Summary())

	opt := adv.GetOneOption(dhcpv6.OptionStatusCode)
	if opt == nil {
		plugin.ServiceOutput = fmt.Sprintf("ERR - No IANA status code in ADVERTISE response")
		return
	}

	resp, _ := opt.(*dhcpv6.OptStatusCode)

	plugin.ExitStatusCode = nagios.StateOKExitCode
	plugin.ServiceOutput = fmt.Sprintf("OK - %s", resp.StatusMessage)
}

func getLinkLocalAddress(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, ifaddr := range addrs {
		if ifaddr, ok := ifaddr.(*net.IPNet); ok {
			if ifaddr.IP.To4() == nil && ifaddr.IP.IsLinkLocalUnicast() {
				return ifaddr.IP, nil
			}
		}
	}

	return nil, nil
}
