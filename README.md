# DNS Water torture attack

This program should be used to perform load tests on DNS servers you control

To build, just run:

```bash
go build watertorture
```

Usage:

```bash
go run watertorture.go -h
Usage of /tmp/go-build097542102/b001/exe/watertorture:
  -count int
    	Count of requests to each DNS server. 0 or less means infinite
  -d	Attack domain nameservers directly instead of public DNS servers
  -delay int
    	Milliseconds between each request (default 1000)
  -f string
    	File with DNS servers
  -s string
    	Initial DNS server. Ignored if a file is specified (default "8.8.8.8")
  -t string
    	Target domain
exit status 2
```

A more exhaustive DNS server list can be found at https://public-dns.info/nameservers.txt

Example usage (just to test, it should not affect anything):

```bash
go run watertorture.go -count 2 -f dnsservers.txt -delay 10000 -t google.com
```

## Bind9 server configuration (for tests)

```bash
domain=example.com
publicip=127.0.0.1
cat << EOF > /etc/default/bind9
RESOLVCONF=no
OPTIONS="-u bind -4"
EOF

cat << EOF > /etc/bind/named.conf.options
options {
    directory "/var/cache/bind";
    dnssec-validation auto;
    listen-on-v6 { any; };
};
EOF

cat << EOF > /etc/bind/named.conf.local
zone "${domain}" {
  type master;
  file "/etc/bind/db.custom-domain.com";
};
EOF

cat << EOF > /etc/bind/db.custom-domain.com
\$TTL    604800
@    IN    SOA     ns1.${domain}. admin.${domain}. (
                  3        ; Serial
             604800        ; Refresh
              86400        ; Retry
            2419200        ; Expire
             604800 )    ; Negative Cache TTL
;
    IN  NS  ns1.${domain}.
    IN  NS  ns2.${domain}.

; name servers - A records
@                       IN  A   ${publicip}
ns1.${domain}.          IN  A   ${publicip}
ns2.${domain}.          IN  A   ${publicip}
EOF

named-checkconf
named-checkzone ${domain} /etc/bind/db.custom-domain.com
systemctl restart bind9

iptables -I INPUT -i eth0 -p udp --dport 53 -j ACCEPT
iptables -I INPUT -i eth0 -p tcp --dport 53 -j ACCEPT
```