package main

import (
	"bytes"
	"fmt"
	"os"

	arg "github.com/alexflint/go-arg"
	nagios "github.com/atc0005/go-nagios"
	log "github.com/sirupsen/logrus"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.PanicLevel)
}

const VRRP_STATE_INIT = 0
const VRRP_STATE_BACKUP = 1
const VRRP_STATE_MASTER = 2
const VRRP_STATE_FAULTY = 3

type Instance struct {
	Name  Name
	State State
}

type Name struct {
	Text string
}

type State struct {
	Code uint
	Text string
}

func (i Instance) String() string {
	return fmt.Sprintf("%s(%s)", i.Name.Text, i.State.Text)
}

func main() {
	var opts struct {
		Address string `arg:"-a,--address" default:"org.keepalived.Vrrp1"`
		Debug   bool   `arg:"--debug"`
	}

	arg.MustParse(&opts)

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	var plugin = nagios.NewPlugin()
	plugin.ExitStatusCode = nagios.StateUNKNOWNExitCode
	defer plugin.ReturnCheckResults()

	conn, err := dbus.SystemBus()
	if err != nil {
		plugin.Errors = append(plugin.Errors, err)
		plugin.ServiceOutput = fmt.Sprintf("UNKNOWN - Error: %s", err)
		return
	}
	defer conn.Close()

	log.Debugf("Recursive scanning for VRRP instances on %s", opts.Address)

	instances, err := scanAll(conn, opts.Address)
	if err != nil {
		plugin.Errors = append(plugin.Errors, err)
		plugin.ServiceOutput = fmt.Sprintf("UNKNOWN - Error: %s", err)
		return
	}

	var faulty []Instance
	var backup []Instance
	var master []Instance
	var unknown []Instance

	for _, path := range instances {
		logger := log.WithField("path", path)
		logger.Debugf("Loading org.keepalived.Vrrp1.Instance")

		instance, err := load(conn, conn.Object(opts.Address, path))
		if err != nil {
			plugin.Errors = append(plugin.Errors, err)
			plugin.ServiceOutput = fmt.Sprintf("UNKNOWN - Error: %s", err)
			return
		}

		logger.Debugf("Status: %v", instance)

		if instance.State.Code == VRRP_STATE_BACKUP {
			backup = append(backup, instance)
		} else if instance.State.Code == VRRP_STATE_MASTER {
			master = append(master, instance)
		} else if instance.State.Code == VRRP_STATE_FAULTY {
			faulty = append(faulty, instance)
		} else {
			unknown = append(unknown, instance)
		}
	}

	log.Debugf("Faulty: %s", faulty)
	log.Debugf("Backup: %s", backup)
	log.Debugf("Master: %s", master)
	log.Debugf("Unknown: %s", unknown)

	perf := []nagios.PerformanceData{
		{
			Label: "faulty",
			Value: fmt.Sprintf("%d", len(faulty)),
		},
		{
			Label: "backup",
			Value: fmt.Sprintf("%d", len(backup)),
		},
		{
			Label: "master",
			Value: fmt.Sprintf("%d", len(master)),
		},
		{
			Label: "unknown",
			Value: fmt.Sprintf("%d", len(unknown)),
		},
	}

	err = plugin.AddPerfData(false, perf...)
	if err != nil {
		log.Debugf("Failed to performance data: %s", err)
		plugin.Errors = append(plugin.Errors, err)
	}

	var buf bytes.Buffer

	if len(faulty) > 0 {
		buf.WriteString(" Faulty: ")

		for idx, instance := range faulty {
			if idx > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(instance.Name.Text)
		}

		plugin.ExitStatusCode = nagios.StateCRITICALExitCode
		plugin.ServiceOutput = fmt.Sprintf("CRITICAL -%s", buf.String())
		return
	}

	if len(unknown) > 0 {
		buf.WriteString(" Unknown: ")

		for idx, instance := range unknown {
			if idx > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(instance.Name.Text)
		}
	}

	if len(master) > 0 {
		buf.WriteString(" Master: ")

		for idx, instance := range master {
			if idx > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(instance.Name.Text)
		}
	}

	if len(backup) > 0 {
		buf.WriteString(" Backup: ")

		for idx, instance := range backup {
			if idx > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(instance.Name.Text)
		}
	}

	plugin.ExitStatusCode = nagios.StateOKExitCode
	plugin.ServiceOutput = fmt.Sprintf("OK -%s", buf.String())
}

func scanAll(conn *dbus.Conn, addr string) (out []dbus.ObjectPath, err error) {
	var scan func(obj dbus.BusObject) error
	scan = func(obj dbus.BusObject) error {
		node, err := introspect.Call(obj)
		if err != nil {
			return err
		}

		if len(node.Interfaces) > 0 {
			for _, iface := range node.Interfaces {
				if iface.Name != "org.keepalived.Vrrp1.Instance" {
					continue
				}

				out = append(out, obj.Path())
			}
		}

		for _, child := range node.Children {
			childPath := string(obj.Path()) + "/" + child.Name
			err := scan(conn.Object(addr, dbus.ObjectPath(childPath)))
			if err != nil {
				return err
			}
		}

		return nil
	}

	bo := conn.Object(addr, "/org/keepalived/Vrrp1/Instance")
	err = scan(bo)
	if err != nil {
		return nil, err
	}

	return out, err
}

func load(conn *dbus.Conn, obj dbus.BusObject) (i Instance, err error) {
	err = obj.StoreProperty("org.keepalived.Vrrp1.Instance.Name", &i.Name)
	if err != nil {
		return i, err
	}

	err = obj.StoreProperty("org.keepalived.Vrrp1.Instance.State", &i.State)
	return i, err
}
