package service

import (
	"time"
)

type ServiceType string

const (
	ServiceTypeDocker  ServiceType = "docker"
	ServiceTypeProcess ServiceType = "process"
	ServiceTypeCompose ServiceType = "compose"
)

type Status string

const (
	StatusRunning   Status = "running"
	StatusStopped   Status = "stopped"
	StatusUnhealthy Status = "unhealthy"
)

type Service struct {
	ID          string
	Name        string
	Type        ServiceType
	Status      Status
	Port        int
	PID         int
	ContainerID string
	Image       string
	Uptime      time.Duration
	StartTime   time.Time
	Project     string
	DependsOn   []string
}

type Store struct {
	services map[string]*Service
}

func NewStore() *Store {
	return &Store{
		services: make(map[string]*Service),
	}
}

func (s *Store) Upsert(service *Service) {
	s.services[service.ID] = service
}

func (s *Store) GetAll() []*Service {
	result := make([]*Service, 0, len(s.services))
	for _, svc := range s.services {
		result = append(result, svc)
	}
	return result
}

func (s *Store) Remove(id string) {
	delete(s.services, id)
}

func (s *Store) Clear() {
	s.services = make(map[string]*Service)
}
