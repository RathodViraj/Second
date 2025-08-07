package handler

import (
	"net/http"
	"second/model"
	"second/repository"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocHandler struct {
	docRepo *repository.DocumentRepo
}

func NewDocHandler(docRepo *repository.DocumentRepo) *DocHandler {
	return &DocHandler{
		docRepo: docRepo,
	}
}

func (h *DocHandler) AddDocument(c *gin.Context) {
	var doc model.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		JSONError(http.StatusBadRequest, c, "Invalid request body")
		return
	}

	missingFields := []string{}
	if doc.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if doc.Content == "" {
		missingFields = append(missingFields, "content")
	}
	if len(missingFields) > 0 {
		JSONError(http.StatusBadRequest, c, "Missing fields: "+strings.Join(missingFields, ", "))
		return
	}

	doc.ID = primitive.NewObjectID()
	doc.Created = time.Now()

	docID, err := h.docRepo.AddDocumentToDB(doc)
	if err != nil {
		JSONError(http.StatusInternalServerError, c, "Failed to add document")
		return
	}

	req := &model.IndexRequest{
		DocID: docID,
		Text:  doc.Content,
	}

	if err := h.docRepo.IndexDoc(req.DocID, req.Text); err != nil {
		JSONError(http.StatusInternalServerError, c, "Faild to index")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "indexed"})
}

func (h *DocHandler) Search(c *gin.Context) {
	word := c.Query("q")
	if word == "" {
		JSONError(http.StatusBadRequest, c, "query param `q` is required")
		return
	}

	page_, exists := c.GetQuery("page")
	if !exists {
		page_ = "1"
	}
	page, err := strconv.ParseInt(page_, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	offset := 10 * (page - 1)

	results, err := h.docRepo.SearchIF_IDF(word, int(offset))
	if err != nil {
		JSONError(http.StatusInternalServerError, c, "search failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h *DocHandler) GetDocumentByID(c *gin.Context) {
	id := c.Param("id")
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		JSONError(http.StatusBadRequest, c, "Invalid ID")
		return
	}

	doc, err := h.docRepo.GetDocumentByID(docID.Hex())
	if err != nil {
		JSONError(http.StatusNotFound, c, "Not found")
	}

	c.JSON(http.StatusOK, doc)
}
