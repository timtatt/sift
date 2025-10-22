package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSummary(t *testing.T) {
	s := NewSummary()

	assert.NotNil(t, s)
	total := s.Total()
	assert.Equal(t, 0, total.Passed)
	assert.Equal(t, 0, total.Failed)
	assert.Equal(t, 0, total.Running)
}

func TestAddPackage(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		wantPass int
		wantFail int
		wantRun  int
	}{
		{
			name:     "pass",
			status:   "pass",
			wantPass: 1,
			wantFail: 0,
			wantRun:  0,
		},
		{
			name:     "fail",
			status:   "fail",
			wantPass: 0,
			wantFail: 1,
			wantRun:  0,
		},
		{
			name:     "run",
			status:   "run",
			wantPass: 0,
			wantFail: 0,
			wantRun:  1,
		},
		{
			name:     "unknown",
			status:   "unknown",
			wantPass: 0,
			wantFail: 0,
			wantRun:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSummary()
			s.AddPackage("pkg1", tt.status)

			total := s.Total()
			assert.Equal(t, tt.wantPass, total.Passed)
			assert.Equal(t, tt.wantFail, total.Failed)
			assert.Equal(t, tt.wantRun, total.Running)
		})
	}

	t.Run("multiple statuses", func(t *testing.T) {
		s := NewSummary()
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg2", "fail")
		s.AddPackage("pkg3", "run")
		s.AddPackage("pkg4", "pass")

		total := s.Total()
		assert.Equal(t, 2, total.Passed)
		assert.Equal(t, 1, total.Failed)
		assert.Equal(t, 1, total.Running)
	})

	t.Run("same package multiple times", func(t *testing.T) {
		s := NewSummary()
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg1", "fail")

		total := s.Total()
		assert.Equal(t, 2, total.Passed)
		assert.Equal(t, 1, total.Failed)
	})
}

func TestPackageSummary(t *testing.T) {
	t.Run("aggregates package data", func(t *testing.T) {
		s := NewSummary()
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg2", "fail")
		s.AddPackage("pkg3", "run")

		pkgSummary := s.PackageSummary()
		assert.Equal(t, 1, pkgSummary.Passed)
		assert.Equal(t, 1, pkgSummary.Failed)
		assert.Equal(t, 1, pkgSummary.Running)
	})

	t.Run("empty summary", func(t *testing.T) {
		s := NewSummary()
		pkgSummary := s.PackageSummary()

		assert.Equal(t, 0, pkgSummary.Passed)
		assert.Equal(t, 0, pkgSummary.Failed)
		assert.Equal(t, 0, pkgSummary.Running)
	})

	t.Run("matches total", func(t *testing.T) {
		s := NewSummary()
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg2", "pass")
		s.AddPackage("pkg3", "fail")
		s.AddPackage("pkg4", "run")

		total := s.Total()
		pkgSummary := s.PackageSummary()

		assert.Equal(t, total.Passed, pkgSummary.Passed)
		assert.Equal(t, total.Failed, pkgSummary.Failed)
		assert.Equal(t, total.Running, pkgSummary.Running)
	})
}

func TestSummary_ComplexScenarios(t *testing.T) {
	t.Run("lifecycle progression", func(t *testing.T) {
		s := NewSummary()
		s.AddPackage("pkg1", "run")
		s.AddPackage("pkg1", "pass")
		s.AddPackage("pkg2", "run")
		s.AddPackage("pkg2", "fail")
		s.AddPackage("pkg3", "run")
		s.AddPackage("pkg3", "pass")
		s.AddPackage("pkg4", "run")

		total := s.Total()
		assert.Equal(t, 2, total.Passed)
		assert.Equal(t, 1, total.Failed)
		assert.Equal(t, 4, total.Running)
	})

	t.Run("multiple packages", func(t *testing.T) {
		s := NewSummary()
		packages := []struct {
			name   string
			status string
		}{
			{name: "pkg1", status: "pass"},
			{name: "pkg2", status: "pass"},
			{name: "pkg3", status: "pass"},
			{name: "pkg4", status: "fail"},
			{name: "pkg5", status: "fail"},
			{name: "pkg6", status: "run"},
		}

		for _, p := range packages {
			s.AddPackage(p.name, p.status)
		}

		total := s.Total()
		assert.Equal(t, 3, total.Passed)
		assert.Equal(t, 2, total.Failed)
		assert.Equal(t, 1, total.Running)

		pkgSummary := s.PackageSummary()
		assert.Equal(t, total, pkgSummary)
	})
}
