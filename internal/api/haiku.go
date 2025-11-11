package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/brianherrera/commits-fall-like-leaves/internal/service/haiku"
	"github.com/gin-gonic/gin"
)

func (api *HaikuAPI) postHaiku(c *gin.Context) {
	var request haiku.HaikuCommitRequest

	// Validate request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("[HAIKU API] error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   InvalidRequest,
			"details": err.Error(),
		})
		return
	}

	// Enforce max commit length
	if len(request.CommitMessage) > MaxCommitLength {
		log.Printf("[HAIKU API] commitMessage exceeds %d characters", MaxCommitLength)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   InvalidRequest,
			"details": fmt.Sprintf("commitMessage exceeds %d characters", MaxCommitLength),
		})
		return
	}

	response, err := api.haikuService.CreateHaiku(c.Request.Context(), request)

	if err != nil {
		if err == haiku.ErrBadHaikuRequest {
			log.Printf("[HAIKU API] bad haiku request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   InvalidRequest,
				"details": err.Error(),
			})
			return
		}

		log.Printf("[HAIKU API] internal server error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
