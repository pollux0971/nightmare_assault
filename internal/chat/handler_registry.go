package chat

import (
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// HandlerRegistry maps ChatFlags to their corresponding handler functions.
// This allows ChatProcessor to dynamically dispatch handlers based on flags.
type HandlerRegistry struct {
	handlers map[ChatFlag]FlagHandler
}

// NewHandlerRegistry creates a new HandlerRegistry with all default handlers registered.
func NewHandlerRegistry() *HandlerRegistry {
	registry := &HandlerRegistry{
		handlers: make(map[ChatFlag]FlagHandler),
	}

	// Register all default handlers
	registry.RegisterHandler(ChatFlagHallucination, handleHallucination)
	registry.RegisterHandler(ChatFlagHostile, handleHostility)
	registry.RegisterHandler(ChatFlagRevelation, handleRevelation)
	registry.RegisterHandler(ChatFlagContradiction, handleContradiction)
	registry.RegisterHandler(ChatFlagPersuasion, handlePersuasion)
	registry.RegisterHandler(ChatFlagLie, handleLie)

	return registry
}

// RegisterHandler registers a handler function for a specific flag.
// This allows custom handlers to be added or default handlers to be overridden.
func (r *HandlerRegistry) RegisterHandler(flag ChatFlag, handler FlagHandler) {
	r.handlers[flag] = handler
}

// GetHandler retrieves the handler function for a specific flag.
// Returns nil if no handler is registered for the flag.
func (r *HandlerRegistry) GetHandler(flag ChatFlag) FlagHandler {
	return r.handlers[flag]
}

// HasHandler checks if a handler is registered for a specific flag.
func (r *HandlerRegistry) HasHandler(flag ChatFlag) bool {
	_, exists := r.handlers[flag]
	return exists
}

// ExecuteHandler executes the handler for a specific flag with the given context.
// Returns an error if no handler is registered for the flag.
func (r *HandlerRegistry) ExecuteHandler(flag ChatFlag, ctx FlagHandlerContext) (FlagHandlerResult, error) {
	handler := r.GetHandler(flag)
	if handler == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("No handler registered for flag: %s", flag.String()),
		}, fmt.Errorf("no handler registered for flag: %s", flag.String())
	}

	result := handler(ctx)
	return result, nil
}

// ExecuteAllHandlers executes handlers for all provided flags in order.
// It accumulates the results from all handlers and returns a combined result.
//
// If any handler fails, it logs the error but continues processing remaining handlers.
// The combined result will have Success=false if any handler failed.
func (r *HandlerRegistry) ExecuteAllHandlers(flags []ChatFlag, ctx FlagHandlerContext) FlagHandlerResult {
	combined := FlagHandlerResult{
		EmotionDelta:   manager.EmotionDelta{},
		NewFacts:       []knowledge.Fact{},
		Contradictions: []knowledge.ContradictionResult{},
		Metadata:       make(map[string]interface{}),
		Success:        true,
	}

	executedHandlers := []string{}

	for _, flag := range flags {
		result, err := r.ExecuteHandler(flag, ctx)
		if err != nil {
			combined.Success = false
			if combined.Error == "" {
				combined.Error = err.Error()
			} else {
				combined.Error += "; " + err.Error()
			}
			continue
		}

		// Accumulate emotion deltas
		combined.EmotionDelta.Trust += result.EmotionDelta.Trust
		combined.EmotionDelta.Fear += result.EmotionDelta.Fear
		combined.EmotionDelta.Stress += result.EmotionDelta.Stress

		// Accumulate new facts
		combined.NewFacts = append(combined.NewFacts, result.NewFacts...)

		// Accumulate contradictions
		combined.Contradictions = append(combined.Contradictions, result.Contradictions...)

		// Merge metadata (with flag prefix to avoid conflicts)
		for k, v := range result.Metadata {
			combined.Metadata[fmt.Sprintf("%s_%s", flag.String(), k)] = v
		}

		if !result.Success {
			combined.Success = false
			if combined.Error == "" {
				combined.Error = result.Error
			} else {
				combined.Error += "; " + result.Error
			}
		}

		executedHandlers = append(executedHandlers, flag.String())
	}

	combined.Metadata["executed_handlers"] = executedHandlers
	combined.Metadata["total_handlers"] = len(flags)

	return combined
}

// GetRegisteredFlags returns a list of all flags that have registered handlers.
func (r *HandlerRegistry) GetRegisteredFlags() []ChatFlag {
	flags := make([]ChatFlag, 0, len(r.handlers))
	for flag := range r.handlers {
		flags = append(flags, flag)
	}
	return flags
}
