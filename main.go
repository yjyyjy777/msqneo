package main

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type TableInfo struct {
	Name      string `json:"name"`
	Rows      int    `json:"rows"`
	DataSize  string `json:"data_size"`
	IndexSize string `json:"index_size"`
}

type MySQLInfo struct {
	Host       string      `json:"host"`
	Port       int         `json:"port"`
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	Connection int         `json:"connection"`
	Tables     []TableInfo `json:"tables"`
}

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.tmpl", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		host := c.PostForm("host")
		port, _ := strconv.Atoi(c.PostForm("port"))
		username := c.PostForm("username")
		password := c.PostForm("password")

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", username, password, host, port)

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer db.Close()
		var a string
		var conn int
		err = db.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(&a, &conn)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		rows, err := db.Query("SELECT TABLE_NAME, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH FROM information_schema.tables WHERE table_schema = 'mdm'")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var tables []TableInfo
		for rows.Next() {
			var name string
			var crows int
			var dataSize, indexSize uint64
			err = rows.Scan(&name, &crows, &dataSize, &indexSize)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			table := TableInfo{
				Name:      name,
				Rows:      crows,
				DataSize:  fmt.Sprintf("%.2f MB", float64(dataSize)/(1024*1024)),
				IndexSize: fmt.Sprintf("%.2f MB", float64(indexSize)/(1024*1024)),
			}
			tables = append(tables, table)
		}

		mysqlInfo := MySQLInfo{
			Host:       host,
			Port:       port,
			Username:   username,
			Password:   password,
			Connection: conn,
			Tables:     tables,
		}

		c.HTML(200, "result.tmpl", gin.H{
			"mysqlInfo": mysqlInfo,
		})
	})

	r.Run(":8080")
}
