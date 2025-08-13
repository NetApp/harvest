package requests

import (
	"net"
	"strconv"
	"time"
)

func worker(address string, ports, results chan int) {
	for p := range ports {
		address := net.JoinHostPort(address, strconv.Itoa(p))
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err != nil {
			results <- p
			continue
		}
		_ = conn.Close()
		results <- 0
	}
}

func CheckFreePorts(address string, ports []int) map[int]struct{} {
	freeports := make(map[int]struct{})

	if len(ports) > 0 {
		// this channel will receive ports to be scanned
		portsWorkers := make(chan int, 10)
		// this channel will receive results of scanning
		results := make(chan int)
		// create a slice to store the results so that they can be sorted later.

		// create a pool of workers
		for range cap(portsWorkers) {
			go worker(address, portsWorkers, results)
		}

		// send ports to be scanned
		go func() {
			for _, p := range ports {
				portsWorkers <- p
			}
		}()

		for range ports {
			port := <-results
			if port != 0 {
				freeports[port] = struct{}{}
			}
		}

		// After all the work has been completed, close the channels
		close(portsWorkers)
		close(results)
	}
	return freeports
}
