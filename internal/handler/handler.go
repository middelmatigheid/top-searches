package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/middelmatigheid/top-searches/internal/models"
)

type Service interface {
	// Storage
	GetTopN(n int64) ([]models.TopSearch, error)

	// Stoplist
	AddStoplistWord(word string)
	RemoveStoplistWord(word string)
	StoplistContains(word string) bool
	GetStoplist() []string
}

type Handler struct {
	Service Service
	Logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: service,
		Logger:  logger,
	}
}

func (h *Handler) GetTopN(c *gin.Context) {
	var top []models.TopSearch
	var err error
	nStr := c.Param("n")
	num, err := strconv.ParseInt(nStr, 10, 64)
	if err != nil {
		if h.Logger != nil {
			h.Logger.Error("Error while parsing N", "error", err, "N", nStr, "url", c.Request.URL)
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error while parsing N", "error": err.Error()})
		return
	}
	top, err = h.Service.GetTopN(num)
	if err != nil {
		if h.Logger != nil {
			h.Logger.Error("Error while processing top searches", "error", err, "url", c.Request.URL)
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error while processing top searches", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Top searches processed successfully", "top-searches": top})
}

func (h *Handler) AddStoplistWord(c *gin.Context) {
	word := c.Param("word")
	h.Service.AddStoplistWord(word)

	c.JSON(http.StatusOK, gin.H{"message": "Word was added"})
}

func (h *Handler) RemoveStoplistWord(c *gin.Context) {
	word := c.Param("word")
	h.Service.RemoveStoplistWord(word)

	c.JSON(http.StatusOK, gin.H{"message": "Word was removed", "word": word})
}

func (h *Handler) GetStoplist(c *gin.Context) {
	stoplist := h.Service.GetStoplist()
	c.JSON(http.StatusOK, gin.H{"message": "Stoplist processed successfully", "stoplist": stoplist})
}
