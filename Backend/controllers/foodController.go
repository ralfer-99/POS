package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Food struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
}

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("food.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	db.AutoMigrate(&Food{})
}

func AddFood(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "File upload failed"})
		return
	}

	filename := fmt.Sprintf("uploads/%s", file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "File save failed"})
		return
	}

	food := Food{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Price:       c.GetFloat64("price"),
		Category:    c.PostForm("category"),
		Image:       file.Filename,
	}

	if err := db.Create(&food).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save food item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food added"})
}

func ListFood(c *gin.Context) {
	var foods []Food
	if err := db.Find(&foods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to retrieve food items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": foods})
}

func RemoveFood(c *gin.Context) {
	id := c.PostForm("id")

	var food Food
	if err := db.First(&food, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Food item not found"})
		return
	}

	imagePath := fmt.Sprintf("uploads/%s", food.Image)
	if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete image file"})
		return
	}

	if err := db.Delete(&food).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete food item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Food removed"})
}

func main() {
	router := gin.Default()

	router.POST("/add-food", AddFood)
	router.GET("/list-food", ListFood)
	router.POST("/remove-food", RemoveFood)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
