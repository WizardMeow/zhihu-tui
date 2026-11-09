package feedtui

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestHotEnterOpensQuestionAnswers(t *testing.T) {
	model := &app{
		mode:   Hot,
		source: &commentTestSource{},
		items:  []feedItem{{id: "7", kind: "question", title: "热榜问题"}},
	}
	model.handleKey(context.Background(), "\r")
	if model.nextView == nil || !model.nextView.loading {
		t.Fatal("Enter on a hot question did not start loading its answer list")
	}
}

type questionTestSource struct {
	commentTestSource
	questionID string
	nextURL    string
	failure    error
}

func (source *questionTestSource) GetQuestionAnswersPage(_ context.Context, questionID, nextURL string, _ int) (map[string]any, error) {
	source.questionID, source.nextURL = questionID, nextURL
	if source.failure != nil {
		return nil, source.failure
	}
	id := "answer-1"
	paging := map[string]any{"next": "next-answers"}
	if nextURL != "" {
		id = "answer-2"
		paging = map[string]any{"is_end": true}
	}
	return map[string]any{
		"data":   []any{map[string]any{"id": id, "content": "<p>完整回答</p>", "thanks_count": 1}},
		"paging": paging,
	}, nil
}

func TestQuestionNavigationAcrossModes(t *testing.T) {
	for _, mode := range []Mode{Following, Recommended, Hot} {
		t.Run(mode.label(), func(t *testing.T) {
			source := &questionTestSource{}
			parent := newApp(source, "self", 140, 30, mode)
			target := map[string]any{
				"id": "original-answer", "type": "answer", "content": "原回答",
				"question": map[string]any{"id": "7", "title": "问题标题"},
			}
			key := keyEvent("a")
			if mode == Hot {
				target = map[string]any{"id": "7", "type": "question", "title": "问题标题"}
				key = "\r"
			}
			item, ok := parseFeedItem(map[string]any{"target": target})
			if !ok {
				t.Fatal("failed to parse source item")
			}
			parent.items = []feedItem{{key: "earlier"}, item}
			parent.index, parent.scroll, parent.sidebarStart = 1, 12, 1
			parent.nextURL = "original-next"
			parent.handleKey(context.Background(), key)
			child := parent.nextView
			if child == nil {
				t.Fatal("question navigation did not create a view")
			}
			parent.nextView = nil
			child.applyFetch(<-child.fetches)
			if source.questionID != "7" || source.nextURL != "" || child.err != nil {
				t.Fatalf("question=%q next=%q err=%v", source.questionID, source.nextURL, child.err)
			}
			if len(child.items) != 1 || child.items[0].body != "完整回答" || child.items[0].title != "问题标题" || child.items[0].questionID != "7" {
				t.Fatalf("answers=%+v", child.items)
			}
			for _, width := range []int{60, 140} {
				child.width = width
				lines, _ := renderApp(child)
				rendered := strings.Join(styledLineTexts(lines), "\n")
				if !strings.Contains(rendered, "Esc 返回") || !strings.Contains(rendered, "问题") {
					t.Fatalf("answer view missing navigation at width %d: %s", width, rendered)
				}
			}
			child.startFetch(context.Background(), false)
			child.applyFetch(<-child.fetches)
			if source.nextURL != "next-answers" || len(child.items) != 2 || !child.end {
				t.Fatalf("pagination next=%q items=%d end=%v", source.nextURL, len(child.items), child.end)
			}
			child.handleKey(context.Background(), "r")
			child.applyFetch(<-child.fetches)
			if len(child.items) != 1 || source.nextURL != "" {
				t.Fatal("answer refresh should replace the list from the first page")
			}
			child.handleKey(context.Background(), keyEscape)
			if child.nextView != parent || parent.index != 1 || parent.scroll != 12 || parent.sidebarStart != 1 || parent.nextURL != "original-next" {
				t.Fatal("return did not preserve original reading state")
			}
		})
	}
}

func TestQuestionNavigationFailureCanReturn(t *testing.T) {
	source := &questionTestSource{failure: errors.New("answer request failed")}
	parent := newApp(source, "self", 80, 24, Hot)
	parent.items = []feedItem{{id: "7", kind: "question", title: "问题"}}
	parent.handleKey(context.Background(), "\r")
	child := parent.nextView
	child.applyFetch(<-child.fetches)
	if child.err == nil || child.loading {
		t.Fatal("answer request failure was not displayed")
	}
	child.handleKey(context.Background(), keyEscape)
	if child.nextView != parent {
		t.Fatal("failed view cannot return to hot list")
	}
}

func TestAnswerWithoutQuestionDoesNotNavigate(t *testing.T) {
	model := &app{items: []feedItem{{kind: "answer", id: "42"}}}
	model.handleKey(context.Background(), "a")
	if model.nextView != nil || !strings.Contains(model.message, "缺少所属问题") {
		t.Fatal("missing question should produce a visible message")
	}
}
