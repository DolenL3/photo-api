package httpcontroller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	swag "github.com/swaggest/swgui/v3"
	"go.uber.org/zap"

	"photoapi/internal/config"
	"photoapi/internal/models"
	photoapi "photoapi/internal/photo-api"
)

// HTTPController is http request controller.
type HTTPController struct {
	router  *gin.Engine
	service *photoapi.PhotoService
	config  *config.HTTPConfig
}

// New returns new HTTPController.
func New(router *gin.Engine, service *photoapi.PhotoService, config *config.HTTPConfig) *HTTPController {
	return &HTTPController{
		router:  router,
		service: service,
		config:  config,
	}
}

// Start starts HTTPController.
func (h *HTTPController) Start() error {
	logger := zap.L()
	currentPath, err := os.Getwd()
	if err != nil {
		logger.Warn("Can't get current path")
	}
	docsPath := currentPath + "/docs/"
	h.router.StaticFile("/static/openapi.json", docsPath+"openapi.json")
	// Retrieves all photos.
	h.router.GET("/photos", h.getPhotos)
	// Uploads new photo.
	h.router.POST("/photos", h.uploadPhoto)
	// Deletes existing photo.
	h.router.DELETE("/photos/:id", h.deletePhoto)
	// Opens API reference.
	h.router.GET("/docs/*any", gin.WrapH(swag.New("PhotoAPI", "/static/openapi.json", "/docs/")))

	logger.Info(fmt.Sprintf("http server is up and running on %s", h.config.Host))
	err = h.router.Run(h.config.Host)
	if err != nil {
		return errors.Wrap(err, "run router")
	}

	return nil
}

func (h *HTTPController) getPhotos(c *gin.Context) {
	photos, err := h.service.Storage.GetPhotos(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dtoPhotos := []photoDTO{}
	for _, photo := range photos {
		dtoPhotos = append(dtoPhotos, photoDTO{
			ID:      photo.ID,
			Bytes:   photo.Bytes,
			Preview: photo.Preview,
		})
	}
	c.JSONP(http.StatusOK, dtoPhotos)
}

func (h *HTTPController) uploadPhoto(c *gin.Context) {
	photo := photoDTO{}
	err := c.ShouldBind(&photo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	res, err := h.service.Storage.UploadPhoto(c.Request.Context(), &models.Photo{
		Bytes: photo.Bytes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSONP(http.StatusOK, photoDTO{
		ID:      res.ID,
		Bytes:   res.Bytes,
		Preview: res.Preview,
	})
}

func (h *HTTPController) deletePhoto(c *gin.Context) {
	photoID := c.Param("id")
	if photoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no id specified"})
		return
	}
	photoUUID, err := uuid.Parse(photoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.service.Storage.DeletePhoto(c.Request.Context(), photoUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
