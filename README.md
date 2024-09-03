# API для управления задачами

Этот проект представляет собой REST API для управления задачами, написанный на Go. Он позволяет создавать, просматривать, обновлять и удалять задачи, каждая из которых имеет заголовок, описание и дату завершения.

## Предварительные требования

- Установлен Go 1.22.6
- Настроенная база данных PostgreSQL

## Начало работы

1. **Клонируйте репозиторий**

   git clone <URL вашего репозитория>


2. Запустите приложение
   go run main.go

3. Введите пароль от PostgreSQL.
   По умолчанию, подключение к базе данных осуществляется с использованием следующих параметров:
   db, err := initdb.NewPostgresConnecction(initdb.ConnectionInfo{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Dbname:   "postgres",
    SSLmode:  "disable",
    Password: password, // ваш пароль

})

4. Создание таблицы
   После успешного соединения с базой данных, будет создана таблица tasks.



Использование API
Создание задачи
Метод: POST /tasks
Описание: Создает новую задачу.
Метод: POST /tasks
Описание: Создать новую задачу.
Запрос:
Заголовки:
Content-Type: application/json
Тело:
{
"title": "string",
"description": "string",
"due_date": "string (RFC3339 format)"
}
Ответ:
Успех (201 Created):
{
"id": "int",
"title": "string",
"description": "string",
"due_date": "string (RFC3339 format)",
"created_at": "string (RFC3339 format)",
"updated_at": "string (RFC3339 format)"
}
Ошибка (400 Bad Request): Неправильный формат данных.
Ошибка (500 Internal Server Error): Проблема на сервере.

