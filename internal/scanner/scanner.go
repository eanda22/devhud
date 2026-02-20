package scanner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/eanda22/devhud/internal/service"
)

type Scanner struct {
	portScanner    *PortScanner
	processScanner *ProcessScanner
	dockerScanner  *DockerScanner
	store          *service.Store
}

// initializes all available discovery methods.
func NewScanner(store *service.Store) (*Scanner, error) {
	dockerScanner, err := NewDockerScanner()
	if err != nil {
		dockerScanner = nil
	}

	return &Scanner{
		portScanner:    NewPortScanner(),
		processScanner: NewProcessScanner(),
		dockerScanner:  dockerScanner,
		store:          store,
	}, nil
}

// discovers all services and updates the store.
func (s *Scanner) Scan(ctx context.Context) error {
	// Save existing services to preserve StartTime for processes
	oldServices := make(map[string]*service.Service)
	for _, svc := range s.store.GetAll() {
		oldServices[svc.ID] = svc
	}

	s.store.Clear()

	if s.dockerScanner != nil {
		_ = s.scanDocker(ctx)
	}

	_ = s.scanPorts(ctx, oldServices)

	_ = s.scanProcesses(ctx, oldServices)

	return nil
}

// discovers running Docker containers.
func (s *Scanner) scanDocker(ctx context.Context) error {
	containers, err := s.dockerScanner.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, c := range containers {
		startTime := time.Unix(c.Created, 0)
		uptime := time.Duration(0)
		if c.State == "running" {
			uptime = time.Since(startTime)
		}

		svc := &service.Service{
			ID:          c.ID,
			Name:        c.Name,
			Type:        service.ServiceTypeDocker,
			ContainerID: c.ID,
			Image:       c.Image,
			DBType:      c.DBType,
			StartTime:   startTime,
			Uptime:      uptime,
		}

		if c.State == "running" {
			svc.Status = service.StatusRunning
		} else {
			svc.Status = service.StatusStopped
		}

		s.store.Upsert(svc)
	}

	return nil
}

// discovers processes listening on common ports.
func (s *Scanner) scanPorts(ctx context.Context, oldServices map[string]*service.Service) error {
	portInfos, err := s.portScanner.ListeningPorts()
	if err != nil {
		return err
	}

	for _, info := range portInfos {
		id := fmt.Sprintf("port-%d", info.Port)
		startTime := time.Now()
		uptime := time.Duration(0)

		// Preserve StartTime if this service existed before
		if old, exists := oldServices[id]; exists {
			startTime = old.StartTime
			uptime = time.Since(startTime)
		}

		svc := &service.Service{
			ID:        id,
			Name:      info.Process,
			Type:      service.ServiceTypeProcess,
			Port:      info.Port,
			Status:    service.StatusRunning,
			StartTime: startTime,
			Uptime:    uptime,
		}

		if info.PID != "" {
			pid, err := strconv.Atoi(info.PID)
			if err == nil {
				svc.PID = pid
			}
		}

		s.store.Upsert(svc)
	}

	return nil
}

// discovers target development processes.
func (s *Scanner) scanProcesses(ctx context.Context, oldServices map[string]*service.Service) error {
	processes, err := s.processScanner.FindProcesses()
	if err != nil {
		return err
	}

	portInfos, _ := s.portScanner.ListeningPorts()
	portPIDs := make(map[string]bool)
	for _, pInfo := range portInfos {
		if pInfo.PID != "" {
			portPIDs[pInfo.PID] = true
		}
	}

	for _, p := range processes {
		if !portPIDs[p.PID] {
			continue
		}

		pid, err := strconv.Atoi(p.PID)
		if err != nil {
			pid = 0
		}

		startTime := time.Now()
		uptime := time.Duration(0)

		// Preserve StartTime if this service existed before
		if old, exists := oldServices[p.PID]; exists {
			startTime = old.StartTime
			uptime = time.Since(startTime)
		}

		svc := &service.Service{
			ID:        p.PID,
			Name:      p.Command,
			Type:      service.ServiceTypeProcess,
			PID:       pid,
			Status:    service.StatusRunning,
			StartTime: startTime,
			Uptime:    uptime,
		}

		s.store.Upsert(svc)
	}

	return nil
}

// closes the scanner.
func (s *Scanner) Close() error {
	if s.dockerScanner != nil {
		return s.dockerScanner.Close()
	}
	return nil
}
