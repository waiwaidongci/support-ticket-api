package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"support-ticket-api/internal/handler"
	"support-ticket-api/internal/middleware"
	"support-ticket-api/internal/model"
	"support-ticket-api/internal/repository"
	"support-ticket-api/internal/service"
)

func New(store repository.Store) *gin.Engine {
	return NewWithClock(store, time.Now)
}

func NewWithClock(store repository.Store, now func() time.Time) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	svc := service.NewWithClock(store, now)
	tickets := handler.NewTicketHandler(svc)

	api := engine.Group("/api/v1")
	{
		api.POST("/tickets", middleware.RequireRole(store, model.RoleCustomer), tickets.Create)
		api.GET("/tickets", middleware.RequireRole(store, model.RoleCustomer, model.RoleAgent, model.RoleSupervisor), tickets.List)
		api.GET("/tickets/:id", middleware.RequireRole(store, model.RoleCustomer, model.RoleAgent, model.RoleSupervisor), tickets.Get)
		api.PATCH("/tickets/:id/assign", middleware.RequireRole(store, model.RoleSupervisor), tickets.Assign)
		api.PATCH("/tickets/:id/claim", middleware.RequireRole(store, model.RoleAgent), tickets.Claim)
		api.PATCH("/tickets/:id/status", middleware.RequireRole(store, model.RoleAgent, model.RoleSupervisor), tickets.UpdateStatus)
		api.PATCH("/tickets/:id/result", middleware.RequireRole(store, model.RoleAgent, model.RoleSupervisor), tickets.UpdateResolution)
		api.PATCH("/tickets/:id/priority", middleware.RequireRole(store, model.RoleSupervisor), tickets.SetPriority)
		api.GET("/tickets/:id/history", middleware.RequireRole(store, model.RoleCustomer, model.RoleAgent, model.RoleSupervisor), tickets.History)
		api.GET("/statistics", middleware.RequireRole(store, model.RoleSupervisor), tickets.Statistics)
	}

	return engine
}
