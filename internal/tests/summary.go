package tests

type TestSummary struct {
	Passed  int
	Failed  int
	Running int
}

type Summary struct {
	packages map[string]TestSummary
	total    TestSummary
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
		ps.Passed += p.Passed
		ps.Failed += p.Failed
		ps.Running += p.Running
	}
	return ps
}
