package scanner

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PortScanner struct {
	commonPorts []int
}

type PortInfo struct {
	Port    int
	Process string
	PID     string
}

func NewPortScanner() *PortScanner {
	return &PortScanner{
		commonPorts: []int{
			3000, 3001, 4000, 5000, 5173, 5432, 6379,
			8000, 8080, 8443, 27017, 9000, 9200,
		},
	}
}

func (ps *PortScanner) ListeningPorts() ([]PortInfo, error) {
	ports, err := ps.scanWithLsof()
	if err == nil && len(ports) > 0 {
		return ports, nil
	}

	return ps.scanWithDial()
}

func (ps *PortScanner) scanWithLsof() ([]PortInfo, error) {
	cmd := exec.Command("lsof", "-i", "-P", "-n", "-sTCP:LISTEN")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ports []PortInfo
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		name := fields[8]
		parts := strings.Split(name, ":")
		if len(parts) != 2 {
			continue
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		ports = append(ports, PortInfo{
			Port:    port,
			Process: fields[0],
			PID:     fields[1],
		})
	}

	return ports, nil
}

func (ps *PortScanner) scanWithDial() ([]PortInfo, error) {
	var ports []PortInfo

	for _, port := range ps.commonPorts {
		address := fmt.Sprintf("localhost:%d", port)
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			ports = append(ports, PortInfo{
				Port:    port,
				Process: "",
				PID:     "",
			})
		}
	}

	return ports, nil
}
