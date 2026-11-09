package feedtui

import (
	"context"
	"strings"
	"testing"
)

type modeTestSource struct {
	commentTestSource
	nextURL  string
	hotLimit int
}

func (source *modeTestSource) GetRecommendedFeed(_ context.Context, nextURL string, _ int) (map[string]any, error) {
	source.nextURL = nextURL
	return map[string]any{
		"data": []any{map[string]any{"target": map[string]any{
			"type": "answer", "id": "42", "question": map[string]any{"id": "7", "title": "推荐问题"},
			"content": "<p>推荐正文</p>", "thanks_count": 1,
		}}},
		"paging": map[string]any{"next": "next-page", "is_end": false},
	}, nil
}

func (source *modeTestSource) GetHotList(_ context.Context, limit int) (map[string]any, error) {
	source.hotLimit = limit
	return map[string]any{"data": []any{map[string]any{
		"question":    map[string]any{"id": "7", "title": "最新热榜", "excerpt": "问题摘要"},
		"detail_text": "100 万热度",
	}}}, nil
}

func TestRecommendedFetchAndPagination(t *testing.T) {
	source := &modeTestSource{}
	model := &app{source: source, mode: Recommended, fetches: make(chan fetchResult, 1)}
	model.startFetch(context.Background(), true)
	model.applyFetch(<-model.fetches)
	if len(model.items) != 1 || model.items[0].body != "推荐正文" || model.end || model.nextURL != "next-page" {
		t.Fatalf("unexpected recommendation state: %+v", model)
	}
	model.startFetch(context.Background(), false)
	model.applyFetch(<-model.fetches)
	if source.nextURL != "next-page" || len(model.items) != 1 {
		t.Fatalf("pagination URL=%q, items=%d", source.nextURL, len(model.items))
	}
}

func TestHotRefreshReplacesRanking(t *testing.T) {
	source := &modeTestSource{}
	model := &app{
		source: source, mode: Hot, fetches: make(chan fetchResult, 1),
		items: []feedItem{{key: "question:old", id: "old", kind: "question", title: "旧榜单"}},
	}
	model.handleKey(context.Background(), "r")
	model.applyFetch(<-model.fetches)
	if source.hotLimit != 50 || !model.end || len(model.items) != 1 {
		t.Fatalf("unexpected hot list state: %+v", model)
	}
	item := model.items[0]
	if item.title != "最新热榜" || item.action != "热榜第 1 名 · 100 万热度" || item.url != "https://www.zhihu.com/question/7" {
		t.Fatalf("item=%+v", item)
	}
}

func TestNormalizeHotListShapes(t *testing.T) {
	question := map[string]any{"id": "7", "title": "问题"}
	for _, raw := range []any{
		question,
		map[string]any{"target": question},
		map[string]any{"question": question},
	} {
		response, err := normalizeHotList(map[string]any{"data": []any{raw}})
		if err != nil {
			t.Fatal(err)
		}
		items := parseFeedItems(asSlice(response["data"]))
		if len(items) != 1 || items[0].kind != "question" || items[0].id != "7" {
			t.Fatalf("items=%+v", items)
		}
	}
	if _, err := normalizeHotList(map[string]any{"data": []any{map[string]any{"title": "缺少 ID"}}}); err == nil {
		t.Fatal("expected invalid question error")
	}
}

func TestModeLabelsInEmptyLayout(t *testing.T) {
	for _, mode := range []Mode{Recommended, Hot} {
		for _, width := range []int{60, 140} {
			model := &app{mode: mode, width: width, height: 24}
			lines, _ := renderApp(model)
			rendered := strings.Join(styledLineTexts(lines), "\n")
			if !strings.Contains(rendered, mode.label()) || strings.Contains(rendered, "知乎关注") {
				t.Fatalf("mode=%d width=%d: %s", mode, width, rendered)
			}
		}
	}
}
