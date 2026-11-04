package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"wraithforge/backend/pkg/database"
)

type Server struct {
	db     *database.DB
	engine *gin.Engine
}

func New(db *database.DB) *Server {
	r := gin.Default()
	s := &Server{db: db, engine: r}

	api := r.Group("/api")
	api.Use(AuthMiddleware())
	api.GET("/flows", s.listFlows)
	api.GET("/flows/:flowId/tasks", s.listTasks)
	api.GET("/tasks/:taskId/subtasks", s.listSubtasks)
	api.GET("/flows/:flowId/agent-logs/stream", s.streamAgentLogs)
	api.GET("/flows/:flowId/results", s.listResults)
	api.GET("/flows/:flowId/cost-summary", s.costSummary)

	return s
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) listFlows(c *gin.Context) {
	flows, err := s.db.ListFlows(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, flows)
}

func (s *Server) listTasks(c *gin.Context) {
	flowID := c.Param("flowId")
	tasks, err := s.db.ListTasksByFlow(c.Request.Context(), flowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (s *Server) listSubtasks(c *gin.Context) {
	taskID := c.Param("taskId")
	subtasks, err := s.db.ListSubtasksByTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subtasks)
}

func (s *Server) listResults(c *gin.Context) {
	flowID := c.Param("flowId")
	results, err := s.db.ListResultsByFlow(c.Request.Context(), flowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (s *Server) costSummary(c *gin.Context) {
	flowID := c.Param("flowId")
	summary, err := s.db.CostSummaryByFlow(c.Request.Context(), flowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (s *Server) streamAgentLogs(c *gin.Context) {
	flowID := c.Param("flowId")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	since := time.Now().Add(-time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			logs, err := s.db.ListAgentLogs(c.Request.Context(), flowID, since)
			if err != nil {
				return
			}
			for _, l := range logs {
				payload, _ := json.Marshal(l)
				c.SSEvent("agentLog", string(payload))
			}
			if len(logs) > 0 {
				since = now
			}
			c.Writer.Flush()
		}
	}
}
