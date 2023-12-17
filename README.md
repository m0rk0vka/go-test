# go-test
Запуск проекта:  
```
docker build . -t test_service:latest
```
```
docker-compose up -d --force-recreate
```
Тестирование через postman:  
Получение списка новостей (по умолчанию выводится только одна новость, т.к. init.sql скрипт заполняет таблицу только тремя новостями, limit=1 поставлен для проверки пагинации):  
URL:  
```
http://localhost:8080/list
```
Чтобы получить другие новости в параметрах нужно выставить значения limit и offset:  
URL:  
```
http://localhost:8080/list?limit=2&offset=1
```
Изменение новости по id:
Header:
```
x-api-key:correct horse battery staple
```
URL:  
```
http://localhost:8080/edit/3
```
Body:  
```
{
    "id": 46,
    "title": "title_(3)",
    "content": "content-(3)",
    "categories": [3, 4, 5]
}
```
Уточнения:  
1. Технически по описании работы ручки на изменение новости по id, я бы использовал метод PATCH.
2. Игнорирую поле id, при изменении новости по id, так как можно попасть на уже существующую id, а так же у нас стоит auto_increment. 
