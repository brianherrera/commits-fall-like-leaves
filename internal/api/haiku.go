package api

import (
	"net/http"

	"github.com/brianherrera/commits-fall-like-leaves/internal/service/haiku"
	"github.com/gin-gonic/gin"
)

func (api *HaikuAPI) postHaiku(c *gin.Context) {
	var request haiku.HaikuCommitRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   InvalidRequest,
			"details": err.Error(),
		})
		return
	}

	response, err := api.haikuService.CreateHaiku(c.Request.Context(), request)

	if err != nil {
		if err == haiku.ErrBadHaikuRequest {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   InvalidRequest,
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
