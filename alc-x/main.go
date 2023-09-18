package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	m "github.com/micarsls/go-learn/alc/models"

	"go.uber.org/zap"
)

var db *sqlx.DB
var logger = zap.Must(zap.NewProduction())

func initDB() error {
	var err error

	connStr := "postgres://admin:123456@localhost/alc?sslmode=disable"
	db, err = sqlx.Open("postgres", connStr)

	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	logger.Info("db connected")
	return nil
}

func initRouter() error {
	router := gin.Default()
	router.POST("/alcs", postAlc)
	router.GET("/alcs", getAlc)
	router.GET("/alcs/:id", getAlcByID)
	router.DELETE("/alcs/:id", deleteAlcByID)
	err := router.Run("localhost:8080")
	if err != nil {
		return err
	}
	return nil
}

func main() {

	if err := initDB(); err != nil {
		logger.Error("failed to init db", zap.String("error", err.Error()))
		return
	}

	defer db.Close()

	if err := initRouter(); err != nil {
		logger.Error("failed to init router", zap.String("error", err.Error()))
		return
	}

	defer logger.Sync()
}

func postAlc(c *gin.Context) {
	var newAlcohol m.Alcohol

	if err := c.BindJSON(&newAlcohol); err != nil {
		logger.Error(
			"error binding alc",
			zap.Error(err),
			zap.String("name", newAlcohol.Name),
			zap.String("description", *newAlcohol.Description),
			zap.Float64("price", newAlcohol.Price),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "error binding"})
		return
	}

	_, err := db.Exec(
		`insert into "alc"("name", "description", "price") values($1, $2, $3)`,
		newAlcohol.Name,
		newAlcohol.Description,
		newAlcohol.Price,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"unique_name\"") {
			logger.Error(
				"already exists",
				zap.Error(err),
				zap.String("name", newAlcohol.Name),
				zap.String("description", *newAlcohol.Description),
				zap.Float64("price", newAlcohol.Price),
			)
			c.JSON(http.StatusConflict, gin.H{"message": "already exists"})
			return
		}
		logger.Error(
			"error creating alc",
			zap.Error(err),
			zap.String("name", newAlcohol.Name),
			zap.String("description", *newAlcohol.Description),
			zap.Float64("price", newAlcohol.Price),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	logger.Info(
		"successfully created",
		zap.String("name", newAlcohol.Name),
		zap.String("description", *newAlcohol.Description),
		zap.Float64("price", newAlcohol.Price),
	)
	c.JSON(http.StatusCreated, gin.H{"message": "successfully created"})
}

func getAlc(c *gin.Context) {

	alcs := []m.Alcohol{}

	err := db.Select(&alcs, `SELECT * FROM alc`)

	if err != nil {
		logger.Error(
			"error getting alc",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if len(alcs) == 0 {
		logger.Error(
			"empty list",
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{"message": "empty list"})
		return
	}

	logger.Info(
		"successfully fetched",
	)
	c.JSON(http.StatusOK, alcs)
}

func getAlcByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		logger.Error(
			"error getting alc by id",
			zap.Error(err),
			zap.Int("id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	result := m.Alcohol{}

	err = db.Get(&result, `SELECT * FROM alc WHERE id=$1;`, id)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error(
				"not found",
				zap.Error(err),
				zap.Int("id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
		logger.Error(
			"error getting alc by id",
			zap.Error(err),
			zap.Int("id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logger.Info(
		"successfully fetched",
		zap.Int("id", id),
	)
	c.JSON(http.StatusOK, result)
}

func deleteAlcByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		logger.Error(
			"error deleting alc",
			zap.Error(err),
			zap.Int("id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	result := m.Alcohol{}

	err = db.Get(&result, `DELETE FROM alc WHERE id=$1 RETURNING *`, id)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error(
				"not found",
				zap.Error(err),
				zap.Int("id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
		logger.Error(
			"error deleting alc",
			zap.Error(err),
			zap.Int("id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// result, err := db.Exec("DELETE from alc WHERE id=$1", id)

	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	// 	return
	// }

	// rows, err := result.RowsAffected()

	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	// 	return
	// }
	// if rows != 1 {
	// 	c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	// 	return
	// }
	logger.Info(
		"successfully deleted",
		zap.Int("id", id),
	)
	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}
