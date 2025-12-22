// Package trinity provides Trinity System routing for different model tiers.
package trinity

import (
	"context"
	"regexp"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ThinkingMiddleware extracts and removes <thinking> tags from responses.
// Story 9-2: Thinking Middleware
type ThinkingMiddleware struct {
	thinkingRegex *regexp.Regexp
}

// NewThinkingMiddleware creates a new ThinkingMiddleware instance.
// Story 9-2 AC1: Implement ThinkingMiddleware structure
func NewThinkingMiddleware() *ThinkingMiddleware {
	// Story 9-2 AC2: Use regex (?s)<thinking>(.*?)</thinking>
	// (?s) enables DOTALL mode where . matches newlines
	// (.*?) is non-greedy matching
	return &ThinkingMiddleware{
		thinkingRegex: regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`),
	}
}

// extractThinking extracts the content inside <thinking> tags.
// Story 9-2 AC2: Extract thinking chain content
func (m *ThinkingMiddleware) extractThinking(content string) string {
	matches := m.thinkingRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1] // Return the captured group (content without tags)
	}
	return "" // No thinking tags found
}

// removeThinkingTags removes <thinking> tags and their content from the response.
// Story 9-2 AC3: Remove thinking tags from response
func (m *ThinkingMiddleware) removeThinkingTags(content string) string {
	return m.thinkingRegex.ReplaceAllString(content, "")
}

// Process wraps a Provider's SendMessage call to extract and remove thinking tags.
// Story 9-2 AC4: Implement Process method
//
// The Process method:
// 1. Calls provider.SendMessage to get the response
// 2. Extracts the <thinking> content and stores it in resp.Metadata["thinking_chain"]
// 3. Removes the <thinking> tags from resp.Content
// 4. Logs the extraction at Debug level
//
// Returns the cleaned response with metadata populated.
func (m *ThinkingMiddleware) Process(ctx context.Context, provider client.Provider, messages []client.Message) (*client.Response, error) {
	// Call the underlying provider
	resp, err := provider.SendMessage(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Extract thinking chain from response
	thinkingChain := m.extractThinking(resp.Content)

	// If thinking tags were found
	if thinkingChain != "" {
		// Store thinking chain in metadata
		if resp.Metadata == nil {
			resp.Metadata = make(map[string]interface{})
		}
		resp.Metadata["thinking_chain"] = thinkingChain

		// Remove thinking tags from content
		resp.Content = m.removeThinkingTags(resp.Content)

		// Log extraction at Debug level (Story 9-2 AC4)
		logger.Debug("ThinkingMiddleware: Extracted thinking chain", map[string]interface{}{
			"chain_length": len(thinkingChain),
		})
	}

	return resp, nil
}
