package service

import (
	"testing"
)

func TestUpsert(t *testing.T) {
	tests := []struct {
		name     string
		services []*Service
		wantLen  int
	}{
		{
			name:     "insert single service",
			services: []*Service{{ID: "a", Name: "nginx"}},
			wantLen:  1,
		},
		{
			name: "insert multiple services",
			services: []*Service{
				{ID: "a", Name: "nginx"},
				{ID: "b", Name: "postgres"},
			},
			wantLen: 2,
		},
		{
			name: "overwrite existing service",
			services: []*Service{
				{ID: "a", Name: "nginx", Status: StatusRunning},
				{ID: "a", Name: "nginx-updated", Status: StatusStopped},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore()
			for _, svc := range tt.services {
				s.Upsert(svc)
			}
			if got := len(s.GetAll()); got != tt.wantLen {
				t.Errorf("GetAll() len = %d, want %d", got, tt.wantLen)
			}
		})
	}
}

func TestUpsertOverwritesFields(t *testing.T) {
	s := NewStore()
	s.Upsert(&Service{ID: "a", Name: "nginx", Status: StatusRunning})
	s.Upsert(&Service{ID: "a", Name: "nginx-v2", Status: StatusStopped})

	all := s.GetAll()
	if all[0].Name != "nginx-v2" {
		t.Errorf("Name = %q, want %q", all[0].Name, "nginx-v2")
	}
	if all[0].Status != StatusStopped {
		t.Errorf("Status = %q, want %q", all[0].Status, StatusStopped)
	}
}

func TestGetAllSorting(t *testing.T) {
	s := NewStore()
	s.Upsert(&Service{ID: "1", Name: "zz-process", Type: ServiceTypeProcess, Status: StatusRunning})
	s.Upsert(&Service{ID: "2", Name: "aa-docker", Type: ServiceTypeDocker, Status: StatusStopped})
	s.Upsert(&Service{ID: "3", Name: "bb-docker", Type: ServiceTypeDocker, Status: StatusRunning})
	s.Upsert(&Service{ID: "4", Name: "cc-docker", Type: ServiceTypeDocker, Status: StatusUnhealthy})

	all := s.GetAll()

	// Docker before process (alphabetical type)
	if all[0].Type != ServiceTypeDocker {
		t.Errorf("first service type = %q, want docker", all[0].Type)
	}

	// Within docker: running < unhealthy < stopped
	expected := []struct {
		name   string
		status Status
	}{
		{"bb-docker", StatusRunning},
		{"cc-docker", StatusUnhealthy},
		{"aa-docker", StatusStopped},
		{"zz-process", StatusRunning},
	}

	for i, want := range expected {
		if all[i].Name != want.name {
			t.Errorf("all[%d].Name = %q, want %q", i, all[i].Name, want.name)
		}
		if all[i].Status != want.status {
			t.Errorf("all[%d].Status = %q, want %q", i, all[i].Status, want.status)
		}
	}
}

func TestGetByType(t *testing.T) {
	s := NewStore()
	s.Upsert(&Service{ID: "1", Name: "nginx", Type: ServiceTypeDocker})
	s.Upsert(&Service{ID: "2", Name: "node", Type: ServiceTypeProcess})
	s.Upsert(&Service{ID: "3", Name: "redis", Type: ServiceTypeDocker})

	docker := s.GetByType(ServiceTypeDocker)
	if len(docker) != 2 {
		t.Fatalf("GetByType(docker) len = %d, want 2", len(docker))
	}
	for _, svc := range docker {
		if svc.Type != ServiceTypeDocker {
			t.Errorf("got type %q, want docker", svc.Type)
		}
	}

	process := s.GetByType(ServiceTypeProcess)
	if len(process) != 1 {
		t.Fatalf("GetByType(process) len = %d, want 1", len(process))
	}

	compose := s.GetByType(ServiceTypeCompose)
	if len(compose) != 0 {
		t.Errorf("GetByType(compose) len = %d, want 0", len(compose))
	}
}

func TestRemove(t *testing.T) {
	s := NewStore()
	s.Upsert(&Service{ID: "a", Name: "nginx"})
	s.Upsert(&Service{ID: "b", Name: "redis"})

	s.Remove("a")
	if len(s.GetAll()) != 1 {
		t.Errorf("after Remove, len = %d, want 1", len(s.GetAll()))
	}

	s.Remove("nonexistent")
	if len(s.GetAll()) != 1 {
		t.Errorf("removing nonexistent changed len to %d, want 1", len(s.GetAll()))
	}
}

func TestClear(t *testing.T) {
	s := NewStore()
	s.Upsert(&Service{ID: "a", Name: "nginx"})
	s.Upsert(&Service{ID: "b", Name: "redis"})

	s.Clear()
	if len(s.GetAll()) != 0 {
		t.Errorf("after Clear, len = %d, want 0", len(s.GetAll()))
	}
}

func TestEmptyStore(t *testing.T) {
	s := NewStore()
	all := s.GetAll()
	if all == nil {
		t.Error("GetAll() on empty store returned nil, want empty slice")
	}
	if len(all) != 0 {
		t.Errorf("GetAll() on empty store len = %d, want 0", len(all))
	}
}
