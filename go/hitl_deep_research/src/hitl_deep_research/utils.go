// Shared helpers used across tools, functions, and workflows.
package hitl_deep_research

import (
	"context"
	"log/slog"

	"github.com/agnt5dev/sdk-go/agnt5"
)

// Tool handlers receive a plain context.Context because that is the signature
// agnt5.ToolHandler requires, but the SDK always passes the run's *agnt5.Context
// through it. These helpers recover it so tool logs land in the run's trace.
//
// The slog fallback matters: without it a handler invoked outside a run (a unit
// test, or a direct agnt5.ToolRegistry.CallTool) would silently drop every log
// line instead of failing loudly or printing somewhere visible.
func logInfo(c context.Context, msg string, kv ...any) {
	if ctx, ok := c.(*agnt5.Context); ok {
		ctx.Logger().Info(msg, kv...)
		return
	}
	slog.InfoContext(c, msg, kv...)
}

func logWarn(c context.Context, msg string, kv ...any) {
	if ctx, ok := c.(*agnt5.Context); ok {
		ctx.Logger().Warn(msg, kv...)
		return
	}
	slog.WarnContext(c, msg, kv...)
}

func logError(c context.Context, msg string, kv ...any) {
	if ctx, ok := c.(*agnt5.Context); ok {
		ctx.Logger().Error(msg, kv...)
		return
	}
	slog.ErrorContext(c, msg, kv...)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
