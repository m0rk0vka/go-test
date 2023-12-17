//go:generate reform
package models

//reform:news
type News struct {
	Id      int    `reform:"id,pk"`
	Title   string `reform:"title"`
	Content string `reform:"content"`
}

//reform:news_categories
type Categories struct {
	NewsId     int `reform:"news_id"`
	CategoryId int `reform:"category_id"`
}
