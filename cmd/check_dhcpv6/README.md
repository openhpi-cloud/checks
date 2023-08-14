# `check_dhcpv6`

Checks a DHCPv6 server by sending a DHCPv6 `SOLICIT` message to a server and expecting a `ADVERTISE` message back. An address is not required.

## Options

<dl>
<dt>

`--interface`, `-i`

</dt>
<dd>

A local interface name were the DHCPv6 SOLICIT request is sent.

**Examples**

Sending a default DHCPv6 request on the `eno1` interface:

```console
check_dhcpv6 --interface eno1
```

</dd>
<dt>

`--address`, `-a`

</dt>
<dd>

The DHCPv6 server address. Defaults to `ff02::1:2`, the multicast address for all local DHCPv6 servers and relays.

When the `address` includes a zone name (interface), the `interface` option is no required.

**Examples**

Directly addressing a specific DHCPv6 server:

```console
check_dhcpv6 --address fe80::b696:91ff:fea5:8bf3%enp7s0
```

Send a request to the default multicast group on a specific interface. This is identical to `-interface enp7s0`:

```console
check_dhcpv6 --address ff02::1:2%enp7s0
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
$ check_dhcpv6 --debug --address ff02::1:2%enp7s0
DEBU[0000] Bind to [fe80::202:c9ff:fe4e:241e%enp7s0]:546
DEBU[0000] Send SOLICIT to [ff02::1:2%enp7s0]:547
DEBU[0000] Message{
  MessageType=SOLICIT
  TransactionID=0x0a1759
  Options: [
    Client ID: DUID-LLT{HWType=Ethernet HWAddr=00:02:c9:4e:24:1e Time=745337393}
    Requested Options: DNS, Domain Search List
    Elapsed Time: 0s
    IANA: IAID=0xc94e241e T1=0s T2=0s Options=[]
  ]
}
DEBU[0000] Received paket: Message{
  MessageType=ADVERTISE
  TransactionID=0x0a1759
  Options: [
    Client ID: DUID-LLT{HWType=Ethernet HWAddr=00:02:c9:4e:24:1e Time=745337393}
    Server ID: DUID-LLT{HWType=Ethernet HWAddr=d0:50:99:df:1d:1c Time=745324685}
    Status Code: {Code=NoAddrsAvail (2); Message=no addresses available}
  ]
}
OK - no addresses available | 'time'=3ms;;;;
```

</dd>
</dl>
