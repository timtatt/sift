package tests

type TestSummary struct {
	Passed  int
	Failed  int
	Skipped int
	Running int
}

type Summary struct {
	packages  map[string]TestSummary
	testTotal TestSummary
}

func NewSummary() *Summary {
	return &Summary{
		packages: make(map[string]TestSummary),
	}
}

func (s *Summary) AddToPackage(pkg string, status string) {
	pkgSummary, ok := s.packages[pkg]
	if !ok {
		pkgSummary = TestSummary{}
	}

	switch status {
	case "error":
		// only increment the failed pkgs count
		// don't increment the total failed tests count
		pkgSummary.Failed++
	case "pass":
		s.testTotal.Passed++
		pkgSummary.Passed++
	case "fail":
		s.testTotal.Failed++
		pkgSummary.Failed++
	case "skip":
		s.testTotal.Skipped++
		pkgSummary.Skipped++
	case "run":
		s.testTotal.Running++
		pkgSummary.Running++
	}

	s.packages[pkg] = pkgSummary
}

func (s *Summary) Total() TestSummary {
	return s.testTotal
}

func (s *Summary) PackageSummary() TestSummary {
	ps := TestSummary{}
	for _, p := range s.packages {
		if p.Running > 0 {
			ps.Running++
		} else if p.Failed > 0 {
			ps.Failed++
		} else {
			ps.Passed++
		}
	}
	return ps
}
