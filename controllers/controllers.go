package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	fiber "github.com/gofiber/fiber/v2"
	reform "gopkg.in/reform.v1"
	postgresql "gopkg.in/reform.v1/dialects/postgresql"

	. "github.com/m0rk0vka/go-test/models"
)

func createConnection() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatalf("Unable to open connection to db.\n %v", err)
	}	

	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to ping db.\n %v", err)
	}

	return db
}

type ListResponse struct {
	Success bool `json:"success"`
	News []NewsWithCategories `json:"news"`
}

func GetNews(c *fiber.Ctx) error {
	log.Println("Get news:")
	sqlDB := createConnection()
	defer sqlDB.Close()

	logger := log.New(os.Stderr, "SQL: ", log.Flags())
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(logger.Printf))

	news, err := db.SelectAllFrom(NewsTable, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	var newsWithCategories []NewsWithCategories
	for _, n := range news {
		log.Printf("For news %v:\n", n.Values()[0])
		newsCategories, _ := db.FindAllFrom(CategoriesView, "news_id", n.Values()[0])
		var categories []int
		for _, nc := range newsCategories{
			log.Printf("Category %v\n", nc.Values()[1])
			categories = append(categories, nc.Values()[1].(int))
		}
		newsWithCategories = append(newsWithCategories, NewsWithCategories{
			Id: n.Values()[0].(int),
			Title: n.Values()[1].(string),
			Content: n.Values()[2].(string),
			Categories: categories,
		})
		log.Println(newsWithCategories)
	}

	lr := ListResponse{
		Success: true,
		News: newsWithCategories,
	}

	log.Printf("listResponse: %v\n", lr)

	return c.JSON(lr)
}

type EditResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

func UpdateNews(c *fiber.Ctx) error {
	log.Println("Edit news by ID:")
	var news NewsWithCategories
	if err := json.Unmarshal([]byte(c.Body()), &news); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	//if err := c.BodyParser(news); err != nil {
	//	c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
	//		"errors": err.Error(),
	//	})
	//	return
	//}
	log.Printf("Body: %v\n", news)
	
	sqlDB := createConnection()
	defer sqlDB.Close()

	logger := log.New(os.Stderr, "SQL: ", log.Flags())
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(logger.Printf))

	tx, err := db.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer func() {
		_ = tx.Rollback()
	}()

	id := c.Params("Id")
	existNews, err := tx.FindByPrimaryKeyFrom(NewsTable, id)
	if err == reform.ErrNoRows {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fmt.Sprintf("No news with id: %v. %v", id, err.Error()),
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// if we left id from the body, we will create a new news
	news.Id = existNews.Values()[0].(int)

	if news.Title == "" {
		news.Title = existNews.Values()[1].(string)
	}

	if news.Content == "" {
		news.Content = existNews.Values()[2].(string)
	}

	newRecord := &News{
		Id: news.Id,
		Title: news.Title,
		Content: news.Content,
	}
	if err := tx.Save(newRecord); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(news.Categories) != 0 {
		tail := fmt.Sprintf("WHERE news_id = %v", news.Id)
		if _, err := tx.DeleteFrom(CategoriesView, tail); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		var categories = []reform.Struct{}
		for _, c := range news.Categories {
			categories = append(categories, &Categories{
				NewsId: news.Id,
				CategoryId: c,
			})
		}
		log.Printf("Categories: %v\n", categories)
		if err := tx.InsertMulti(categories...); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}
	
	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	res := EditResponse{
		Success: true,
		Message: "Successifully edit the news",
	}
	return c.JSON(res)
}