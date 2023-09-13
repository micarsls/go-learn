package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	m "github.com/micarsls/go-learn/alc/models"
)

var db *sql.DB
var alc = make(map[int]m.Alcohol)

func initDB() error {
	var err error

	connStr := "postgres://admin:123456@localhost/alc?sslmode=disable"
	db, err = sql.Open("postgres", connStr)

	if err != nil {
		fmt.Println("db error: ", err)
		return err
	}

	if err = db.Ping(); err != nil {
		fmt.Println("db error: ", err)
		return err
	}

	fmt.Println("db connected")
	return nil
}

func initRouter() error {
	router := gin.Default()
	router.POST("/alcs", postAlc)
	router.GET("/alcs", getAlc)
	router.GET("/alcs/:id", getAlcByID)
	router.DELETE("/alcs/:id", deleteAlcByID)
	err := router.Run("localhost:8080")
	return err
}

func main() {
	if err := initDB(); err != nil {
		fmt.Println("failed to init db")
		return
	}

	defer db.Close()

	if err := initRouter(); err != nil {
		fmt.Println("failed to init router")
		return
	}
}

func postAlc(c *gin.Context) {
	var newAlcohol m.Alcohol

	if err := c.BindJSON(&newAlcohol); err != nil {
		println(err.Error())
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
			c.JSON(http.StatusConflict, gin.H{"message": "already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successfully created"})
}

func getAlc(c *gin.Context) {

	alcs := []m.Alcohol{}

	rows, err := db.Query(`SELECT * FROM alc`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var name, description string
		var price float64

		err := rows.Scan(&id, &name, &description, &price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		alcs = append(alcs, m.Alcohol{ID: id, Name: name, Description: &description, Price: price})
	}

	if len(alcs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "empty list"})
		return
	}

	c.JSON(http.StatusOK, alcs)
}

func getAlcByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	result := m.Alcohol{}

	row := db.QueryRow(`SELECT * FROM alc WHERE id=$1;`, id)
	err = row.Scan(&result.ID, &result.Name, &result.Description, &result.Price)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func deleteAlcByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	result := m.Alcohol{}

	row := db.QueryRow(`DELETE FROM alc WHERE id=$1 RETURNING *`, id)

	err = row.Scan(&result.ID, &result.Name, &result.Description, &result.Price)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
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

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}
