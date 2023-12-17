package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	validator "github.com/go-playground/validator/v10"
	fiber "github.com/gofiber/fiber/v2"
	reform "gopkg.in/reform.v1"
	postgresql "gopkg.in/reform.v1/dialects/postgresql"

	"github.com/m0rk0vka/go-test/models"
)


type (
	NewsWithCategories struct {
		Id int `json:"id"`
		Title string `json:"title" validate:"min=3,max=50"`
		Content string `json:"content" validate:"min=5,max=500"`
		Categories []int `json:"categories" validate:"unique"`
	}

	ErrorResponse struct {
        Error       bool
        FailedField string
        Tag         string
        Value       interface{}
    }

    XValidator struct {
        validator *validator.Validate
    }

	GlobalErrorHandlerResp struct {
        Success bool   `json:"success"`
        Message string `json:"message"`
    }
	
	ListResponse struct {
		Success bool `json:"success"`
		News []NewsWithCategories `json:"news"`
	}

 	EditResponse struct {
		Success bool `json:"success"`
		Message string `json:"message"`
	}
)

var validate = validator.New()

func (v XValidator) Validate(data interface{}) []ErrorResponse {
    validationErrors := []ErrorResponse{}

    errs := validate.Struct(data)
    if errs != nil {
        for _, err := range errs.(validator.ValidationErrors) {
            // In this case data object is actually holding the User struct
            var elem ErrorResponse

            elem.FailedField = err.Field() // Export struct field name
            elem.Tag = err.Tag()           // Export struct tag
            elem.Value = err.Value()       // Export field value
            elem.Error = true

            validationErrors = append(validationErrors, elem)
        }
    }

    return validationErrors
}

func createConnection() *reform.DB {
	sqlDB, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatalf("Unable to open connection to db.\n %v", err)
	}	

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Unable to ping db.\n %v", err)
	}

	logger := log.New(os.Stderr, "SQL: ", log.Flags())
	db := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(logger.Printf))

	return db
}


func GetNews(c *fiber.Ctx) error {
	log.Println("Get news:")
	db := createConnection()

	limitString := c.Query("limit", "1")
	offsetString := c.Query("offset", "0")
	limit, err := strconv.Atoi(limitString)
	if err != nil {
		return &fiber.Error{
			Code: fiber.ErrBadRequest.Code,
			Message: fmt.Sprintf(
				"limit parameter %s is invalid, should be positive integer",
				limitString,
			),
		}
	}
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		return &fiber.Error{
			Code: fiber.ErrBadRequest.Code,
			Message: fmt.Sprintf(
				"offset parameter %s is invalid, should be positive integer",
				offsetString,
			),
		}
	}

	tail := fmt.Sprintf("ORDER BY id LIMIT %v OFFSET %v", limit, offset)
	news, err := db.SelectAllFrom(models.NewsTable, tail)
	if err != nil {
		return &fiber.Error{
			Code: fiber.ErrInternalServerError.Code,
			Message: err.Error(),
		}
	}
	
	var newsWithCategories []NewsWithCategories
	for _, n := range news {
		log.Printf("For news %v:\n", n.Values()[0])
		newsCategories, _ := db.FindAllFrom(models.CategoriesView, "news_id", n.Values()[0])
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



func UpdateNews(c *fiber.Ctx) error {
	log.Println("Edit news by ID:")
	var news NewsWithCategories
	if err := json.Unmarshal([]byte(c.Body()), &news); err != nil {
		return &fiber.Error{
			Code: fiber.ErrBadRequest.Code,
			Message: err.Error(),
		}
	}
	//if err := c.BodyParser(news); err != nil {
	//	c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
	//		"errors": err.Error(),
	//	})
	//	return
	//}
	log.Printf("Body: %v\n", news)

	myValidator := &XValidator{
        validator: validate,
    }
	// Validation
	if errs := myValidator.Validate(news); len(errs) > 0 && errs[0].Error {
		errMsgs := make([]string, 0)

		for _, err := range errs {
			errMsgs = append(errMsgs, fmt.Sprintf(
				"[%s]: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}

		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: strings.Join(errMsgs, " and "),
		}
	}
	
	db := createConnection()

	tx, err := db.Begin()
	if err != nil {
		return &fiber.Error{
			Code: fiber.ErrInternalServerError.Code,
			Message: err.Error(),
		}
	}
	defer func() {
		_ = tx.Rollback()
	}()

	id := c.Params("Id")
	existNews, err := tx.FindByPrimaryKeyFrom(models.NewsTable, id)
	if err == reform.ErrNoRows {
		return &fiber.Error{
			Code: fiber.ErrBadRequest.Code,
			Message: fmt.Sprintf("No news with id: %v. %v", id, err.Error()),
		}
	}
	if err != nil {
		return &fiber.Error{
			Code: fiber.ErrInternalServerError.Code,
			Message: err.Error(),
		}
	}

	// if we left id from the body, we will create a new news
	news.Id = existNews.Values()[0].(int)

	if news.Title == "" {
		news.Title = existNews.Values()[1].(string)
	}

	if news.Content == "" {
		news.Content = existNews.Values()[2].(string)
	}

	newRecord := &models.News{
		Id: news.Id,
		Title: news.Title,
		Content: news.Content,
	}
	if err := tx.Save(newRecord); err != nil {
		return &fiber.Error{
			Code: fiber.ErrInternalServerError.Code,
			Message: err.Error(),
		}
	}

	if len(news.Categories) != 0 {
		tail := fmt.Sprintf("WHERE news_id = %v", news.Id)
		if _, err := tx.DeleteFrom(models.CategoriesView, tail); err != nil {
			return &fiber.Error{
				Code: fiber.ErrInternalServerError.Code,
				Message: err.Error(),
			}
		}
		var categories = []reform.Struct{}
		for _, c := range news.Categories {
			categories = append(categories, &models.Categories{
				NewsId: news.Id,
				CategoryId: c,
			})
		}
		log.Printf("Categories: %v\n", categories)
		if err := tx.InsertMulti(categories...); err != nil {
			return &fiber.Error{
				Code: fiber.ErrInternalServerError.Code,
				Message: err.Error(),
			}
		}
	}
	
	if err := tx.Commit(); err != nil {
		return &fiber.Error{
			Code: fiber.ErrInternalServerError.Code,
			Message: err.Error(),
		}
	}

	res := EditResponse{
		Success: true,
		Message: "Successifully edit the news",
	}
	return c.JSON(res)
}