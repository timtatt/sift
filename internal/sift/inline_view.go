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

	stack := newTestStack()
	var lastPackage string

	for _, test := range m.testManager.GetTests {
		summary.AddToPackage(test.Ref.Package, test.Status)

		if test.Ref.Package != lastPackage {
			if lastPackage != "" {
				vb.AddLine()
			}

			style := styleSecondary
			prefix := ""

			if test.Ref.Test == "" {
				style = style.Foreground(colorMutedRed)
				prefix = style.Foreground(colorRed).Render("! ")
			}

			vb.Add(prefix + style.Render(test.Ref.Package))
			vb.AddLine()
			lastPackage = test.Ref.Package
		}

		if test.Ref.Test != "" {
			statusIcon := m.getStatusIcon(test.Status)

			prefixTest := stack.PopUntilPrefix(test.Ref.Test)
			testName, _ := strings.CutPrefix(test.Ref.Test, prefixTest)

			indentLevel := stack.Len()
			indent := getIndentWithBars(indentLevel)

			elapsed := ""
			if test.Status != "run" {
				elapsed = styleSecondary.Render(
					formatDuration(test.Elapsed),
				)
			}

			vb.Add(fmt.Sprintf("%s%s %s %s", indent, statusIcon, testName, elapsed))
			vb.AddLine()
		} else {
			for _, logEntry := range m.testManager.GetLogs(test.Ref) {

				prettifiedLog := prettifyLogEntry(logEntry, styleLog)
				vb.Add(fmt.Sprintf("%s", prettifiedLog))
				vb.AddLine()
			}

		}

		stack.Push(test.Ref.Test)
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
