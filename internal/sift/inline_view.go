package sift

import (
	"fmt"
	"strings"

	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/viewbuilder"
)

func (m *siftModel) inlineView() string {
	if !m.started {
		return styleSecondary.Render("Waiting for test results...")
	}

	vb := viewbuilder.New()
	summary := tests.NewSummary()

	for _, test := range m.testManager.GetTests {
		summary.AddPackage(test.Ref.Package, test.Status)

		var statusIcon string
		switch test.Status {
		case "skip":
			statusIcon = styleSkip.Render("\u23ED")
		case "run":
			statusIcon = styleProgress.Render("\u2022")
		case "fail":
			statusIcon = styleCross.Render("\u00D7")
		case "pass":
			statusIcon = styleTick.Render("\u2713")
		}

		indentLevel := getIndentLevel(test.Ref.Test)
		indent := strings.Repeat("  ", indentLevel)
		testName := getDisplayName(test.Ref.Test)

		elapsed := ""
		if test.Status != "run" {
			elapsed = styleSecondary.Render(
				fmt.Sprintf(" (%.2fs)", test.Elapsed.Seconds()),
			)
		}

		vb.Add(fmt.Sprintf("%s%s %s%s", indent, statusIcon, testName, elapsed))
		vb.AddLine()
	}

	vb.AddLine()
	vb.Add(m.summaryView(summary))

	if !m.endTime.IsZero() {
		vb.AddLine()
		vb.AddLine()
		total := summary.Total()
		if total.Failed > 0 {
			vb.Add(styleOutcomeFail.Render("FAILED"))
		} else {
			vb.Add(styleOutcomePass.Render("PASSED"))
		}
		vb.AddLine()
	}

	return vb.String()
}
