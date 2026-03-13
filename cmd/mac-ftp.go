package main

import (
	"log"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/mdlayher/arp"
)

func main() {

	iface, err := net.InterfaceByName("wlp4s0")
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}

	for _, value := range addrs {
		ipv4Addr, ipv4Net, err := net.ParseCIDR(value.String())
		if err != nil {
			log.Fatal(err)
		}

		if ipv4Addr.To4() != nil {
			var n int = 0
			for _, b := range ipv4Net.Mask {
				n *= 256
				n += int(^b)
			}

			for i := 1; i < n; i++ {
				ip, err := netip.ParseAddr(
					net.IPv4(
						ipv4Net.IP[0],
						ipv4Net.IP[1],
						ipv4Net.IP[2],
						byte(i),
					).String())
				if err != nil {
					log.Fatal(err)
				}

				wg.Go(func() {
					c, err := arp.Dial(iface)
					if err != nil {
						log.Fatal(err)
					}
					defer c.Close()

					if err := c.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
						log.Fatal(err)
					}

					mac, err := c.Resolve(ip)
					if err != nil {
						log.Printf("%s -> %s\n", ip, err.Error())
						return
					}

					log.Printf("%s -> %s\n", ip, mac)
				})
			}

			wg.Wait()
		}
	}
}
