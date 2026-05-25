package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/middelmatigheid/top-searches/producer/internal/models"
)

type Service interface {
	SendSearch(request models.Search) error
}

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Search(c *gin.Context) {
	var request models.Search
	if err := c.ShouldBindJSON(&request); err != nil {
		if h.logger != nil {
			h.logger.Error("Error while reading request's body", "error", err)
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error while reading request's body", "error": err.Error()})
		return
	}
	err := h.service.SendSearch(request)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Error while searching", "error", err)
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error while searching", "error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Search message sent"})
	}
}
