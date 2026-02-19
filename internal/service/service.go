package service

import (
	"sort"
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

// manages discovered services.
func NewStore() *Store {
	return &Store{
		services: make(map[string]*Service),
	}
}

// stores a service by ID.
func (s *Store) Upsert(service *Service) {
	s.services[service.ID] = service
}

// returns services sorted by type, status, and name.
func (s *Store) GetAll() []*Service {
	result := make([]*Service, 0, len(s.services))
	for _, svc := range s.services {
		result = append(result, svc)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type < result[j].Type
		}
		statusOrder := map[Status]int{
			StatusRunning:   0,
			StatusUnhealthy: 1,
			StatusStopped:   2,
		}
		if statusOrder[result[i].Status] != statusOrder[result[j].Status] {
			return statusOrder[result[i].Status] < statusOrder[result[j].Status]
		}
		return result[i].Name < result[j].Name
	})
	return result
}

// deletes a service by ID.
func (s *Store) Remove(id string) {
	delete(s.services, id)
}

// empties the store.
func (s *Store) Clear() {
	s.services = make(map[string]*Service)
}
