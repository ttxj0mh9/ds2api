package sse

import "fmt"

// LineResult is the normalized parse result for one DeepSeek SSE line.
type LineResult struct {
	Parsed        bool
	Stop          bool
	ContentFilter bool
	ErrorMessage  string
	Parts         []ContentPart
	NextType      string
	PromptTokens  int
	OutputTokens  int
}

// ParseDeepSeekContentLine centralizes one-line DeepSeek SSE parsing for both
// streaming and non-streaming handlers.
func ParseDeepSeekContentLine(raw []byte, thinkingEnabled bool, currentType string) LineResult {
	chunk, done, parsed := ParseDeepSeekSSELine(raw)
	if !parsed {
		return LineResult{NextType: currentType}
	}
	promptTokens, outputTokens := extractAccumulatedTokenUsage(chunk)
	if done {
		return LineResult{Parsed: true, Stop: true, NextType: currentType, PromptTokens: promptTokens, OutputTokens: outputTokens}
	}
	if errObj, hasErr := chunk["error"]; hasErr {
		return LineResult{
			Parsed:       true,
			Stop:         true,
			ErrorMessage: fmt.Sprintf("%v", errObj),
			NextType:     currentType,
			PromptTokens: promptTokens,
			OutputTokens: outputTokens,
		}
	}
	if code, _ := chunk["code"].(string); code == "content_filter" {
		return LineResult{
			Parsed:        true,
			Stop:          true,
			ContentFilter: true,
			NextType:      currentType,
			PromptTokens:  promptTokens,
			OutputTokens:  outputTokens,
		}
	}
	if hasContentFilterStatus(chunk) {
		return LineResult{
			Parsed:        true,
			Stop:          true,
			ContentFilter: true,
			NextType:      currentType,
			PromptTokens:  promptTokens,
			OutputTokens:  outputTokens,
		}
	}
	parts, finished, nextType := ParseSSEChunkForContent(chunk, thinkingEnabled, currentType)
	parts = filterLeakedContentFilterParts(parts)
	return LineResult{
		Parsed:       true,
		Stop:         finished,
		Parts:        parts,
		NextType:     nextType,
		PromptTokens: promptTokens,
		OutputTokens: outputTokens,
	}
}
