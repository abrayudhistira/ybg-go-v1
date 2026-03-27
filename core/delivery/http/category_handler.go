package http

import (
	"net/http"
	"strconv"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	uc usecase.CategoryUsecase
}

func NewCategoryHandler(uc usecase.CategoryUsecase) *CategoryHandler { return &CategoryHandler{uc: uc} }

func (h *CategoryHandler) Create(c *gin.Context) {
	var cat entity.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.uc.CreateCategory(&cat)
	c.JSON(http.StatusCreated, gin.H{"data": cat})
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, _ := h.uc.GetAllCategories()
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.uc.DeleteCategory(uint(id))
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}
