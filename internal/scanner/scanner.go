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

func NewScanner(store *service.Store) (*Scanner, error) {
	dockerScanner, err := NewDockerScanner()
	if err != nil {
		fmt.Printf("Warning: Docker not available: %v\n", err)
	}

	return &Scanner{
		portScanner:    NewPortScanner(),
		processScanner: NewProcessScanner(),
		dockerScanner:  dockerScanner,
		store:          store,
	}, nil
}

func (s *Scanner) Scan(ctx context.Context) error {
	s.store.Clear()

	if s.dockerScanner != nil {
		if err := s.scanDocker(ctx); err != nil {
			fmt.Printf("Docker scan error: %v\n", err)
		}
	}

	if err := s.scanPorts(ctx); err != nil {
		fmt.Printf("Port scan error: %v\n", err)
	}

	if err := s.scanProcesses(ctx); err != nil {
		fmt.Printf("Process scan error: %v\n", err)
	}

	return nil
}

func (s *Scanner) scanDocker(ctx context.Context) error {
	containers, err := s.dockerScanner.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, c := range containers {
		svc := &service.Service{
			ID:          c.ID,
			Name:        c.Name,
			Type:        service.ServiceTypeDocker,
			ContainerID: c.ID,
			Image:       c.Image,
			StartTime:   time.Now(),
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

func (s *Scanner) scanPorts(ctx context.Context) error {
	portInfos, err := s.portScanner.ListeningPorts()
	if err != nil {
		return err
	}

	for _, info := range portInfos {
		svc := &service.Service{
			ID:        fmt.Sprintf("port-%d", info.Port),
			Name:      info.Process,
			Type:      service.ServiceTypeProcess,
			Port:      info.Port,
			Status:    service.StatusRunning,
			StartTime: time.Now(),
		}

		if info.PID != "" {
			pid, _ := strconv.Atoi(info.PID)
			svc.PID = pid
		}

		s.store.Upsert(svc)
	}

	return nil
}

func (s *Scanner) scanProcesses(ctx context.Context) error {
	processes, err := s.processScanner.FindProcesses()
	if err != nil {
		return err
	}

	for _, p := range processes {
		pid, _ := strconv.Atoi(p.PID)
		svc := &service.Service{
			ID:        p.PID,
			Name:      p.Command,
			Type:      service.ServiceTypeProcess,
			PID:       pid,
			Status:    service.StatusRunning,
			StartTime: time.Now(),
		}

		s.store.Upsert(svc)
	}

	return nil
}

func (s *Scanner) Close() error {
	if s.dockerScanner != nil {
		return s.dockerScanner.Close()
	}
	return nil
}
