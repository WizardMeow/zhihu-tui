package feedtui

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (model *app) openQuestionAnswers(ctx context.Context) {
	if len(model.items) == 0 {
		return
	}
	item := model.items[model.index]
	question := item
	switch item.kind {
	case "question":
	case "answer":
		if item.questionID == "" {
			model.setMessage("这条回答缺少所属问题信息", 3*time.Second)
			return
		}
		question = feedItem{id: item.questionID, kind: "question", title: item.title}
	default:
		model.setMessage("当前内容没有对应的问题回答列表", 3*time.Second)
		return
	}
	if model.question != nil && model.question.id == question.id {
		model.setMessage("已在该问题的回答列表中", 2*time.Second)
		return
	}
	child := newApp(model.source, model.currentUserID, model.width, model.height, model.mode)
	child.question = &question
	child.parent = model
	child.zenMode = model.zenMode
	model.nextView = child
	child.startFetch(ctx, true)
}

func normalizeQuestionAnswers(response map[string]any, question feedItem) (map[string]any, error) {
	data, ok := response["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("question answers response is missing data")
	}
	items := make([]any, 0, len(data))
	for index, raw := range data {
		answer, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(toString(answer["id"])) == "" {
			return nil, fmt.Errorf("answer %d is missing id", index+1)
		}
		answer["type"] = "answer"
		answer["question"] = map[string]any{"id": question.id, "title": question.title}
		items = append(items, map[string]any{"target": answer})
	}
	return map[string]any{"data": items, "paging": response["paging"]}, nil
}
