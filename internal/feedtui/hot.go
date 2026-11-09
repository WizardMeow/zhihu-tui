package feedtui

import (
	"fmt"
	"strings"

	"github.com/JimChengLin/zhihu-tui/internal/display"
)

// The hot endpoint returns questions either directly or under target/question.
func normalizeHotList(response map[string]any) (map[string]any, error) {
	data, ok := response["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("hot list response is missing data")
	}
	items := make([]any, 0, len(data))
	for index, value := range data {
		raw := mapValue(value)
		target := mapValue(raw["target"])
		if len(target) == 0 {
			target = mapValue(raw["question"])
		}
		if len(target) == 0 {
			target = raw
		}
		if strings.TrimSpace(toString(target["id"])) == "" || strings.TrimSpace(toString(target["title"])) == "" {
			return nil, fmt.Errorf("hot list item %d is missing question id or title", index+1)
		}
		question := make(map[string]any, len(target)+1)
		for key, value := range target {
			question[key] = value
		}
		question["type"] = "question"
		heat := toString(raw["detail_text"])
		if heat == "" {
			if pv, present := mapValue(raw["reaction"])["pv"]; present {
				heat = display.FormatCount(pv) + " 热度"
			}
		}
		action := fmt.Sprintf("热榜第 %d 名", index+1)
		if heat != "" {
			action += " · " + heat
		}
		items = append(items, map[string]any{"target": question, "action_text": action})
	}
	return map[string]any{
		"data":   items,
		"paging": map[string]any{"is_end": true},
	}, nil
}
