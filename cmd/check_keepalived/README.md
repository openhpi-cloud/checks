# `check_keepalived`

Uses D-Bus to check the state of a local `keepalived` process. The `keepalived` must be configured to have D-Bus enabled:

```text
global_defs {
  enable_dbus
}
```

## Options

<dl>
<dt>

`--address`, `-a`

</dt>
<dd>

D-Bus address to scan for VRRP instances. Defaults to `org.keepalived.Vrrp1`.

**Examples**

Check default local `keepalived` with two VRRP instances:

```console
check_dhcpv6 --interface eno1
```

</dd>
<dt>

`--debug`

</dt>
<dd>

Turn on debug logging.

> [!NOTE]
> Do not enable this when running actual checks.

**Example**

```session
$ check_keepalived --address org.keepalived.Vrrp2 --debug
DEBU[0000] Recursive scanning for VRRP instances on org.keepalived.Vrrp2
DEBU[0000] Loading org.keepalived.Vrrp1.Instance         path=/org/keepalived/Vrrp1/Instance/enp7s0/51/IPv4
DEBU[0000] Status: vrrp51(Master)                        path=/org/keepalived/Vrrp1/Instance/enp7s0/51/IPv4
DEBU[0000] Loading org.keepalived.Vrrp1.Instance         path=/org/keepalived/Vrrp1/Instance/enp7s0/52/IPv4
DEBU[0000] Status: vrrp52(Master)                        path=/org/keepalived/Vrrp1/Instance/enp7s0/52/IPv4
DEBU[0000] Faulty: []
DEBU[0000] Backup: []
DEBU[0000] Master: [vrrp51(Master) vrrp52(Master)]
DEBU[0000] Unknown: []
OK - Master: vrrp51,vrrp52 | 'backup'=0;;;; 'faulty'=0;;;; 'master'=2;;;; 'time'=3ms;;;; 'unknown'=0;;;;
```

</dd>
</dl>
