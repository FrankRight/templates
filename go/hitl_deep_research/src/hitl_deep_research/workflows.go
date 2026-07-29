// Deep research workflow with human-in-the-loop approval.
//
// A research pipeline using 3 specialized agents with a durable HITL gate:
//  1. Scoping Agent  - analyzes topic and creates a research plan
//  2. [HITL]         - user approval of research plan (approve/edit/reject)
//  3. Research Agent - conducts systematic research using tools
//  4. Writing Agent  - synthesizes findings into a comprehensive report
//
// The workflow pauses for user approval after the research plan is
// generated. ctx.AskUser's first call returns a *WaitingForUserInputError
// that must be propagated as this workflow's own return error so the
// runtime can suspend the run; on resume, the same call returns the user's
// answer directly instead of pausing again.
//
// Replay safety: on resume this function body re-runs from the top. Every
// line above an AskUser call executes a second time — only the AskUser call
// itself short-circuits, returning the recorded answer. Anything with a side
// effect that sits before a pause must therefore go through agnt5.Task or
// agnt5.Step, which memoize their result and skip the work on the replay
// pass. Unlike the Python SDK's ctx._is_replay, the Go SDK exposes no replay
// flag, so there is no way to guard a bare statement — wrapping it is the
// only option. The Logger calls below are the one deliberate exception: they
// duplicate on resume, which is noise rather than a correctness problem.
package hitl_deep_research

import (
	"github.com/agnt5dev/sdk-go/agnt5"
)

type DeepResearchInput struct {
	Message string `json:"message"`
}

type DeepResearchOutput struct {
	Status       string `json:"status"`
	Topic        string `json:"topic"`
	ResearchPlan string `json:"research_plan,omitempty"`
	Report       string `json:"report,omitempty"`
	Message      string `json:"message,omitempty"`
}

func DeepResearchWorkflow(ctx *agnt5.Context, in DeepResearchInput) (DeepResearchOutput, error) {
	topic := in.Message
	ctx.Logger().Info("Deep research workflow started", "topic", truncate(topic, 100))

	// Stage 1: Planning.
	//
	// agnt5.Task rather than agnt5.Step: both checkpoint the result, but Task
	// also emits function.started/function.completed around it, so each stage
	// renders as a Function node in Studio the way the Python and TypeScript
	// versions of this template do.
	researchPlan, err := agnt5.Task(ctx, "plan_research",
		PlanResearchInput{Topic: topic}, PlanResearch)
	if err != nil {
		return DeepResearchOutput{}, err
	}
	ctx.Logger().Info("Research plan created")

	// Stage 2: HITL — wait for user approval of the research plan. This
	// pause is durable — it survives worker restarts and can wait
	// indefinitely for user input.
	decision, err := ctx.AskUser(agnt5.UserInputRequest{
		Prompt: "Please review the research plan below:\n\n---\n" + researchPlan + "\n---\n\nDo you approve this research plan to proceed with research?",
		Type:   agnt5.HITLSelect,
		Options: []agnt5.HITLOption{
			{Label: "Approve Plan", Value: "approve"},
			{Label: "Edit Plan", Value: "edit"},
			{Label: "Reject", Value: "reject"},
		},
	})
	if err != nil {
		return DeepResearchOutput{}, err
	}
	ctx.Logger().Info("Decision received", "decision", decision)

	if decision == "reject" {
		ctx.Logger().Info("Research plan rejected by user")
		return DeepResearchOutput{
			Status:       "rejected",
			Topic:        topic,
			ResearchPlan: researchPlan,
			Message:      "Research plan was rejected. Please start a new session with updated requirements.",
		}, nil
	}

	if decision == "edit" {
		ctx.Logger().Info("User chose to edit the research plan")
		edited, err := ctx.AskUser(agnt5.UserInputRequest{
			Prompt: "Please provide your edited research plan.\n\nHere is the original plan for reference:\n---\n" + researchPlan + "\n---\n\nPaste your revised plan below:",
			Type:   agnt5.HITLText,
		})
		if err != nil {
			return DeepResearchOutput{}, err
		}
		researchPlan = edited
		ctx.Logger().Info("Received edited research plan from user")
	}

	ctx.Logger().Info("Research plan approved, proceeding to research phase")

	// Stage 3: Research
	researchFindings, err := agnt5.Task(ctx, "conduct_research",
		ConductResearchInput{Topic: topic, ResearchPlan: researchPlan}, ConductResearch)
	if err != nil {
		return DeepResearchOutput{}, err
	}
	ctx.Logger().Info("Research findings gathered")

	// Stage 4: Write Report
	finalReport, err := agnt5.Task(ctx, "write_report",
		WriteReportInput{
			Topic:            topic,
			ResearchPlan:     researchPlan,
			ResearchFindings: researchFindings,
		}, WriteReport)
	if err != nil {
		return DeepResearchOutput{}, err
	}

	ctx.Logger().Info("Research completed successfully")
	return DeepResearchOutput{Status: "completed", Topic: topic, Report: finalReport}, nil
}
