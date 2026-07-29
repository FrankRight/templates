// Travel booking workflow.
package travel_booking_customer_service

import (
	"context"

	"github.com/agnt5dev/sdk-go/agnt5"
)

type TravelBookingInput struct {
	Message string `json:"message"`
}

type TravelBookingOutput struct {
	Status string `json:"status"`
	Output string `json:"output"`
}

// TravelBookingWorkflow is a chat-based travel booking workflow. Conversation
// history is kept in session-scoped memory so the agent never asks for
// information the user already gave earlier in the session.
func TravelBookingWorkflow(ctx *agnt5.Context, in TravelBookingInput) (TravelBookingOutput, error) {
	ctx.Logger().Info("Travel booking workflow", "message", truncate(in.Message, 100))

	conversation := ctx.Memory().Conversation()
	history, err := conversation.Messages(ctx)
	if err != nil {
		return TravelBookingOutput{}, err
	}
	messages := make([]agnt5.Message, len(history))
	for i, m := range history {
		messages[i] = agnt5.Message{Role: agnt5.MessageRole(m.Role), Content: m.Content}
	}

	// Wrapped in a Step so the model call is checkpointed. A trip-planning turn
	// makes up to 8 model round-trips and three SerpAPI calls; without this a
	// worker restart before the conversation.Append calls below would re-run
	// and re-bill the entire turn.
	result, err := agnt5.Step(ctx, "travel_booking_agent_turn", func(context.Context) (agnt5.AgentResult, error) {
		return TravelBookingAgent.Run(ctx, agnt5.AgentInput{Messages: messages, Message: in.Message})
	})
	if err != nil {
		return TravelBookingOutput{}, err
	}

	if err := conversation.Append(ctx, agnt5.MemoryMessage{Role: "user", Content: in.Message}); err != nil {
		return TravelBookingOutput{}, err
	}
	if err := conversation.Append(ctx, agnt5.MemoryMessage{Role: "assistant", Content: result.Response}); err != nil {
		return TravelBookingOutput{}, err
	}

	ctx.Logger().Info("Travel booking agent completed")
	return TravelBookingOutput{Status: "completed", Output: result.Response}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
