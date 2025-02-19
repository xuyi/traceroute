package main

import (
	"flag"
	"fmt"
	"github.com/aeden/traceroute"
	"net"
	// ip dat
	"github.com/wangtuanjie/ip17mon"
	"os"
	"os/exec"
	"path/filepath"
)

func printHop(hop traceroute.TracerouteHop) {
	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}

	addrLoc, err := ip17mon.Find(fmt.Sprintf("%v", addr))
	if err != nil {
        fmt.Println("ip17mon error:", err)
        return
    }

	if hop.Success {
		fmt.Printf("%-3d %v (%v) %s  %v\n", hop.TTL, hostOrAddr, addr, LocString(addrLoc), hop.ElapsedTime)
	} else {
		fmt.Printf("%-3d *\n", hop.TTL)
	}
}

func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}

// init ip data
func init() {
	exec_path, _ := exec.LookPath(os.Args[0])
    if err := ip17mon.Init(filepath.Dir(exec_path) + "/17monipdb.dat"); err != nil {
        panic(err)
    }
}

func LocString(loc *ip17mon.LocationInfo) string {
	return fmt.Sprintf("%s %s %s %s", loc.Country, loc.Region, loc.City, loc.Isp)
}

func main() {
	var m = flag.Int("m", traceroute.DEFAULT_MAX_HOPS, `Set the max time-to-live (max number of hops) used in outgoing probe packets (default is 64)`)
	var q = flag.Int("q", 1, `Set the number of probes per "ttl" to nqueries (default is one probe).`)

	flag.Parse()
	host := flag.Arg(0)
	options := traceroute.TracerouteOptions{}
	options.SetRetries(*q - 1)
	options.SetMaxHops(*m + 1)

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return
	}

	loc, err := ip17mon.Find(fmt.Sprintf("%v", ipAddr))
	if err != nil {
        fmt.Println("ip17mon error:", err)
        return
    }

	fmt.Printf("traceroute to %v (%v) %s, %v hops max, %v byte packets Location\n", host, ipAddr, LocString(loc), options.MaxHops(), options.PacketSize())

	c := make(chan traceroute.TracerouteHop, 0)
	go func() {
		for {
			hop, ok := <-c
			if !ok {
				fmt.Println()
				return
			}
			printHop(hop)
		}
	}()

	_, err = traceroute.Traceroute(host, &options, c)
	if err != nil {
		fmt.Printf("Error: ", err)
	}
}
