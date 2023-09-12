package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	m "github.com/micarsls/go-learn/alc/models"
)

var alc = make(map[int]m.Alcohol)

func main() {
	router := gin.Default()
	router.POST("/alcs", postAlc)
	router.GET("/alcs", getAlc)
	router.GET("/alcs/:id", getAlcByID)
	router.DELETE("/alcs/:id", deleteAlcByID)
	router.Run("localhost:8080")
}

func postAlc(c *gin.Context) {
	var newAlcohol m.Alcohol

	if err := c.BindJSON(&newAlcohol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "error creating"})
		return
	}

	if _, check := alc[newAlcohol.ID]; check {
		c.JSON(http.StatusConflict, gin.H{"message": "already exists"})
		return
	}

	alc[newAlcohol.ID] = newAlcohol

	c.JSON(http.StatusCreated, gin.H{"message": "successfully created"})
}

func getAlc(c *gin.Context) {

	if len(alc) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "empty list"})
		return
	}

	mapAlcohol := make([]m.Alcohol, 0, len(alc))

	for _, value := range alc {
		mapAlcohol = append(mapAlcohol, value)
	}

	c.JSON(http.StatusOK, mapAlcohol)
}

func getAlcByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if alc, check := alc[id]; check {
		c.JSON(http.StatusOK, alc)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
}

func deleteAlcByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if _, check := alc[id]; check {
		delete(alc, id)
		c.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
}
