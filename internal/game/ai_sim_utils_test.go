package game

import (
	"strings"
)

func setupSimulationGame() (*Game, *SystemContext) {
	g := setupTestGame(); g.EventLog = []LogEntry{}
	sysCtx := g.GetContext(); sysCtx.Log = func(msg string, category LogCategory) { g.EventLog = append(g.EventLog, LogEntry{Text: msg, Category: category}) }
	return g, sysCtx
}
func hasLog(g *Game, keyword string) bool {
	keyword = strings.ToLower(keyword)
	for _, entry := range g.EventLog { if strings.Contains(strings.ToLower(entry.Text), keyword) { return true } }
	return false
}
func logMessage(ctx *SystemContext, msg string, category LogCategory) { if ctx != nil && ctx.Log != nil { ctx.Log(msg, category) } }
