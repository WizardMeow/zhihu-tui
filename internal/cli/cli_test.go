package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpAdvertisesInteractiveModes(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(--help)=%d, stderr=%q", code, stderr.String())
	}
	for _, command := range []string{
		"zhihu feed --tui",
		"zhihu notifications --monitor",
	} {
		if !strings.Contains(stdout.String(), command) {
			t.Fatalf("root help does not contain %q:\n%s", command, stdout.String())
		}
	}
	lines := strings.Split(stdout.String(), "\n")
	descriptionColumns := make([]int, 0, 2)
	for _, line := range lines {
		for _, description := range []string{"Browse your following feed", "Monitor new notifications"} {
			if column := strings.Index(line, description); column >= 0 {
				descriptionColumns = append(descriptionColumns, column)
			}
		}
	}
	if len(descriptionColumns) != 2 || descriptionColumns[0] != descriptionColumns[1] {
		t.Fatalf("interactive descriptions are not aligned: columns=%v\n%s", descriptionColumns, stdout.String())
	}
}

func TestRootHelpOmitsRemovedCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(--help)=%d, stderr=%q", code, stderr.String())
	}
	for _, removed := range []string{"search, hot, question", "feeds, topic"} {
		if strings.Contains(stdout.String(), removed) {
			t.Fatalf("root help still contains removed command list fragment %q:\n%s", removed, stdout.String())
		}
	}
}

func TestFeedTUIDispatchesBeforeRegularFeedOptions(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"feed", "--tui"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Run(feed --tui)=%d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "requires an interactive terminal") {
		t.Fatalf("feed --tui was not dispatched to the TUI: %q", stderr.String())
	}
}

func TestRecommendedAndHotTUIDispatch(t *testing.T) {
	for _, command := range []string{"recommend", "hot"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run([]string{command, "--tui"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "requires an interactive terminal") {
				t.Fatalf("code=%d stderr=%q", code, stderr.String())
			}
			stdout.Reset()
			stderr.Reset()
			if code := Run([]string{command, "--help"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), command+" --tui") {
				t.Fatalf("help code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}
