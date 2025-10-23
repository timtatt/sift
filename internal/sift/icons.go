package sift

func getStatusIcon(status string) string {
	switch status {
	case "skip":
		return styleSkip.Render("\u23ED")
	case "run":
		return styleProgress.Render("\u2022")
	case "fail":
		return styleCross.Render("\u00D7")
	case "pass":
		return styleTick.Render("\u2713")
	default:
		return ""
	}
}
