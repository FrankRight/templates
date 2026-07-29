// Deep Research Agent — AGNT5 worker.
package main

import (
	"context"
	"log"
	"os"

	"github.com/agnt5dev/sdk-go/agnt5"

	hitl_deep_research "hitl-deep-research/src/hitl_deep_research"
)

const serviceName = "deep-research"

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Printf("Starting %s worker...", serviceName)

	model := agnt5.NewOpenAIModel(agnt5.OpenAIConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o-mini",
	})

	if err := hitl_deep_research.NewAgents(model); err != nil {
		log.Fatal(err)
	}
	log.Println("Worker created successfully with a 3-agent research pipeline:")
	log.Println("  - 3 specialized agents: Scoping, Research, Writing")
	log.Println("  - 3 pipeline functions: plan_research, conduct_research, write_report")
	log.Println("  - 2 research tools: fetch_webpage_tool, wikipedia_search_tool")
	log.Println("  - 1 main workflow: deep_research_workflow")

	worker := agnt5.NewWorker(serviceName,
		agnt5.WithServiceVersion("1.0.0"),
	)

	// The Go SDK has no auto-register equivalent: every agent, function, tool,
	// and workflow has to be listed here explicitly.
	must(agnt5.RegisterAgent(worker, hitl_deep_research.ScopingAgent))
	must(agnt5.RegisterAgent(worker, hitl_deep_research.ResearchAgent))
	must(agnt5.RegisterAgent(worker, hitl_deep_research.WritingAgent))

	must(agnt5.RegisterTool(worker, hitl_deep_research.FetchWebpageTool))
	must(agnt5.RegisterTool(worker, hitl_deep_research.WikipediaSearchTool))

	// Each stage is one LLM call plus (for conduct_research) external HTTP, so
	// all three get the same retry policy: 3 attempts, exponential backoff from
	// 1s up to a 30s ceiling.
	must(agnt5.RegisterFunction(worker, "plan_research", hitl_deep_research.PlanResearch,
		agnt5.WithRetry(3, 1000, 30000),
		agnt5.WithBackoff("exponential", 2.0),
	))
	must(agnt5.RegisterFunction(worker, "conduct_research", hitl_deep_research.ConductResearch,
		agnt5.WithRetry(3, 1000, 30000),
		agnt5.WithBackoff("exponential", 2.0),
	))
	must(agnt5.RegisterFunction(worker, "write_report", hitl_deep_research.WriteReport,
		agnt5.WithRetry(3, 1000, 30000),
		agnt5.WithBackoff("exponential", 2.0),
	))

	must(agnt5.RegisterWorkflow(worker, "deep_research_workflow", hitl_deep_research.DeepResearchWorkflow))

	log.Println("Starting worker and registering with coordinator...")
	if err := worker.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
