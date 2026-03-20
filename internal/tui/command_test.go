package tui

import (
	"testing"

	"github.com/eanda22/devhud/internal/service"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		category  string
		action    string
		target    string
		verbFirst bool
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
		{
			name:     "category alias c",
			input:    "c list",
			category: "containers",
			action:   "list",
		},
		{
			name:     "category alias db",
			input:    "db browse",
			category: "containers",
			action:   "browse",
		},
		{
			name:     "category alias p",
			input:    "p kill",
			category: "processes",
			action:   "kill",
		},
		{
			name:     "full category name",
			input:    "containers list",
			category: "containers",
			action:   "list",
		},
		{
			name:     "category with target",
			input:    "containers stop nginx",
			category: "containers",
			action:   "stop",
			target:   "nginx",
		},
		{
			name:     "category with multi-word target",
			input:    "containers stop my app",
			category: "containers",
			action:   "stop",
			target:   "my app",
		},
		{
			name:      "short-form verb stop",
			input:     "stop nginx",
			category:  "containers",
			action:    "stop",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:      "short-form verb kill",
			input:     "kill node",
			category:  "processes",
			action:    "kill",
			target:    "node",
			verbFirst: true,
		},
		{
			name:      "short-form verb browse",
			input:     "browse mydb",
			category:  "containers",
			action:    "browse",
			target:    "mydb",
			verbFirst: true,
		},
		{
			name:      "short-form verb no target",
			input:     "stop",
			category:  "containers",
			action:    "stop",
			verbFirst: true,
		},
		{
			name:   "unknown command",
			input:  "foobar",
			action: "foobar",
		},
		{
			name:   "unknown command with args",
			input:  "foobar baz qux",
			action: "foobar",
			target: "baz qux",
		},
		{
			name:     "case insensitive category",
			input:    "Containers List",
			category: "containers",
			action:   "list",
		},
		{
			name:      "case insensitive verb",
			input:     "STOP nginx",
			category:  "containers",
			action:    "stop",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:     "category only no action",
			input:    "containers",
			category: "containers",
		},
		{
			name:      "shell as verb",
			input:     "shell nginx",
			category:  "containers",
			action:    "shell",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:      "delete as verb",
			input:     "delete nginx",
			category:  "containers",
			action:    "delete",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:      "single-letter b alias",
			input:     "b mydb",
			category:  "containers",
			action:    "b",
			target:    "mydb",
			verbFirst: true,
		},
		{
			name:      "single-letter r alias",
			input:     "r nginx",
			category:  "containers",
			action:    "r",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:      "single-letter l alias",
			input:     "l app",
			category:  "containers",
			action:    "l",
			target:    "app",
			verbFirst: true,
		},
		{
			name:      "single-letter d alias",
			input:     "d nginx",
			category:  "containers",
			action:    "d",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:      "single-letter i alias",
			input:     "i nginx",
			category:  "containers",
			action:    "i",
			target:    "nginx",
			verbFirst: true,
		},
		{
			name:      "single-letter s toggle",
			input:     "s nginx",
			action:    "s",
			target:    "nginx",
			verbFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommand(tt.input)
			if got.Category != tt.category {
				t.Errorf("Category = %q, want %q", got.Category, tt.category)
			}
			if got.Action != tt.action {
				t.Errorf("Action = %q, want %q", got.Action, tt.action)
			}
			if got.Target != tt.target {
				t.Errorf("Target = %q, want %q", got.Target, tt.target)
			}
			if got.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", got.Raw, tt.input)
			}
			if got.VerbFirst != tt.verbFirst {
				t.Errorf("VerbFirst = %v, want %v", got.VerbFirst, tt.verbFirst)
			}
		})
	}
}

func TestFilterPrefix(t *testing.T) {
	candidates := []string{"containers", "processes", "help", "quit"}

	tests := []struct {
		name   string
		prefix string
		want   int
	}{
		{"empty prefix matches all", "", 4},
		{"matches con", "con", 1},
		{"matches p", "p", 1},
		{"no match", "xyz", 0},
		{"case insensitive", "CON", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterPrefix(candidates, tt.prefix)
			if len(got) != tt.want {
				t.Errorf("filterPrefix(%q) len = %d, want %d", tt.prefix, len(got), tt.want)
			}
		})
	}
}

func testApp(services ...*service.Service) *App {
	store := service.NewStore()
	for _, svc := range services {
		store.Upsert(svc)
	}
	return &App{services: store}
}

func TestCompletions(t *testing.T) {
	nginx := &service.Service{ID: "1", Name: "nginx", Type: service.ServiceTypeDocker, Status: service.StatusRunning}
	redis := &service.Service{ID: "2", Name: "redis", Type: service.ServiceTypeDocker, Status: service.StatusStopped}
	postgres := &service.Service{ID: "3", Name: "postgres", Type: service.ServiceTypeDocker, Status: service.StatusRunning, DBType: "postgres"}
	node := &service.Service{ID: "4", Name: "node-app", Type: service.ServiceTypeProcess, Status: service.StatusRunning}

	app := testApp(nginx, redis, postgres, node)

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty shows verbs and categories",
			input: "",
			want:  []string{"stop", "start", "restart", "kill", "logs", "inspect", "shell", "delete", "browse", "containers", "processes", "help", "quit"},
		},
		{
			name:  "partial st matches stop and start",
			input: "st",
			want:  []string{"stop", "start"},
		},
		{
			name:  "stop space shows running containers only",
			input: "stop ",
			want:  []string{"nginx", "postgres"},
		},
		{
			name:  "browse space shows DB services only",
			input: "browse ",
			want:  []string{"postgres"},
		},
		{
			name:  "kill space shows running processes only",
			input: "kill ",
			want:  []string{"node-app"},
		},
		{
			name:  "logs space shows all services",
			input: "logs ",
			want:  []string{"nginx", "postgres", "redis", "node-app"},
		},
		{
			name:  "containers space shows actions",
			input: "containers ",
			want:  []string{"list", "stop", "start", "restart", "logs", "inspect", "shell", "delete", "browse"},
		},
		{
			name:  "processes space shows actions",
			input: "p ",
			want:  []string{"list", "kill"},
		},
		{
			name:  "containers stop space shows running containers",
			input: "containers stop ",
			want:  []string{"nginx", "postgres"},
		},
		{
			name:  "containers partial action",
			input: "containers st",
			want:  []string{"stop", "start"},
		},
		{
			name:  "start space shows stopped containers",
			input: "start ",
			want:  []string{"redis"},
		},
		{
			name:  "s space shows toggleable services",
			input: "s ",
			want:  []string{"nginx", "postgres", "redis", "node-app"},
		},
		{
			name:  "b space shows DB services",
			input: "b ",
			want:  []string{"postgres"},
		},
		{
			name:  "stop ng filters to nginx",
			input: "stop ng",
			want:  []string{"nginx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.completions(tt.input)
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("completions(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveServiceWithFilter(t *testing.T) {
	redisDocker := &service.Service{ID: "1", Name: "redis", Type: service.ServiceTypeDocker, Status: service.StatusRunning}
	redisProc := &service.Service{ID: "2", Name: "redis-cli", Type: service.ServiceTypeProcess, Status: service.StatusRunning}

	app := testApp(redisDocker, redisProc)

	processFilter := func(svc *service.Service) bool {
		return svc.Type == service.ServiceTypeProcess
	}

	t.Run("kill redis-cli matches process", func(t *testing.T) {
		svc, errMsg := app.resolveService("redis-cli", processFilter)
		if svc == nil || svc.ID != "2" {
			t.Errorf("expected redis-cli process, got svc=%v err=%q", svc, errMsg)
		}
	})

	t.Run("kill redis does not match docker container", func(t *testing.T) {
		svc, errMsg := app.resolveService("redis", processFilter)
		if svc != nil {
			t.Errorf("expected nil for docker redis with process filter, got %v", svc.Name)
		}
		if errMsg != errFiltered {
			t.Errorf("expected filtered error, got %q", errMsg)
		}
	})

	t.Run("no filter matches exact name", func(t *testing.T) {
		svc, errMsg := app.resolveService("redis", nil)
		if svc == nil || svc.Name != "redis" {
			t.Errorf("expected redis, got svc=%v err=%q", svc, errMsg)
		}
	})
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
