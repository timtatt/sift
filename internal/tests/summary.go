package tests

type TestSummary struct {
	Passed  int
	Failed  int
	Skipped int
	Running int
}

type Summary struct {
	packages map[string]TestSummary
	total    TestSummary
}

func NewSummary() *Summary {
	return &Summary{
		packages: make(map[string]TestSummary),
	}
}

func (s *Summary) AddPackage(pkg string, status string) {
	ps, ok := s.packages[pkg]
	if !ok {
		ps = TestSummary{}
	}

	switch status {
	case "pass":
		s.total.Passed++
		ps.Passed++
	case "fail":
		s.total.Failed++
		ps.Failed++
	case "skip":
		s.total.Skipped++
		ps.Skipped++
	case "run":
		s.total.Running++
		ps.Running++
	}

	s.packages[pkg] = ps
}

func (s *Summary) Total() TestSummary {
	return s.total
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
