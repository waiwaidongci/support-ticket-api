package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"support-ticket-api/internal/middleware"
	"support-ticket-api/internal/model"
	"support-ticket-api/internal/pkg/httputil"
	"support-ticket-api/internal/repository"
	"support-ticket-api/internal/service"
)

type TicketHandler struct {
	service *service.Service
}

func NewTicketHandler(service *service.Service) *TicketHandler {
	return &TicketHandler{service: service}
}

func (h *TicketHandler) Create(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	var input service.CreateTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	ticket, err := h.service.CreateTicket(c.Request.Context(), user, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	filter, err := parseListFilter(c)
	if err != nil {
		httputil.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.service.ListTickets(c.Request.Context(), user, filter)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "count": len(items)})
}

func (h *TicketHandler) Get(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	ticket, err := h.service.GetTicket(c.Request.Context(), user, ticketID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) Assign(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	var body struct {
		AssigneeID int64 `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.AssigneeID <= 0 {
		httputil.RespondError(c, http.StatusBadRequest, "assignee_id is required")
		return
	}
	ticket, err := h.service.AssignTicket(c.Request.Context(), user, ticketID, body.AssigneeID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) Claim(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	ticket, err := h.service.ClaimTicket(c.Request.Context(), user, ticketID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	var input service.UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.TicketID = ticketID
	ticket, err := h.service.UpdateStatus(c.Request.Context(), user, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateResolution(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	var input service.UpdateResolutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.TicketID = ticketID
	ticket, err := h.service.UpdateResolution(c.Request.Context(), user, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) SetPriority(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	var input service.SetPriorityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.TicketID = ticketID
	ticket, err := h.service.SetPriority(c.Request.Context(), user, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) History(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	ticketID, ok := parsePathID(c)
	if !ok {
		return
	}
	history, err := h.service.History(c.Request.Context(), user, ticketID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history, "count": len(history)})
}

func (h *TicketHandler) Statistics(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		httputil.RespondError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	stats, err := h.service.Statistics(c.Request.Context(), user)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func parsePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httputil.RespondError(c, http.StatusBadRequest, "invalid ticket id")
		return 0, false
	}
	return id, true
}

func parseListFilter(c *gin.Context) (model.TicketListFilter, error) {
	filter := model.TicketListFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
	}
	if filter.Status != "" && !model.ValidStatus(filter.Status) {
		return filter, errors.New("unsupported status filter")
	}
	if filter.Priority != "" && !model.ValidPriority(filter.Priority) {
		return filter, errors.New("unsupported priority filter")
	}
	if rawAssignee := c.Query("assignee_id"); rawAssignee != "" {
		assigneeID, err := strconv.ParseInt(rawAssignee, 10, 64)
		if err != nil || assigneeID <= 0 {
			return filter, errors.New("invalid assignee_id filter")
		}
		filter.AssigneeID = &assigneeID
	}
	return filter, nil
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		httputil.RespondError(c, http.StatusNotFound, "ticket not found")
	case errors.Is(err, repository.ErrAlreadyAssigned):
		httputil.RespondError(c, http.StatusConflict, "ticket already assigned")
	case errors.Is(err, repository.ErrConflict):
		httputil.RespondError(c, http.StatusConflict, "ticket state changed, please retry")
	case errors.Is(err, service.ErrForbidden):
		httputil.RespondError(c, http.StatusForbidden, "permission denied")
	case errors.Is(err, service.ErrInvalidInput):
		httputil.RespondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidTransition):
		httputil.RespondError(c, http.StatusUnprocessableEntity, err.Error())
	default:
		httputil.RespondError(c, http.StatusInternalServerError, "internal server error")
	}
}
