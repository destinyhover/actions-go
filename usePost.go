/*
Пакет работает на 2х таблицах в PostrgreSQL БД сервер.

Имена таблиц следующие:
  - Users
  - Userdata

Определение таблиц в бд сервере:

	CREATE TABLE Users (
	ID SERIAL,
	Username VARCHAR(100) PRIMARY KEY
	);

	CREATE TABLE Userdata (
	UserID Int NOT NULL,
	Name VARCHAR(100),
	Description VARCHAR(200)
	);
*/
package post05

// BUG(1): Функция ListUsers() работает не так, как ожидалось
// BUG(2): Фу-ция AddUser() is too slow

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

/*
Этот  блок глобальных переменных содержит данные для соеденения с сервером БД

	Hostname: is IP or hosname
	Port: is the server port
	Username: of the db user
	Password: of the db user
	Database: the name of db
*/
var (
	Hostname = ""
	Port     = 0
	Username = ""
	Password = ""
	Database = ""
)

// Структура Userdata предназначчена для хранения полных данных
// о пользователе из таблицы Userdata и имени пользователя из таблицы Users
type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

// openConnection() предназначена для открытия соединения с постгрес
// чтобы его могли использовать другие ф-ции пакета
func openConnection() (*sql.DB, error) {
	// connection string
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	// open database
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Функция exists() возвращается ID of username
// -1 if the user does not exist
func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := -1
	statement := fmt.Sprintf(`SELECT "id" FROM "users" where username = '%s'`, username)
	rows, err := db.Query(statement)
	if err != nil {
		fmt.Println("Rows", err)
		return -1
	}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("Scan", err)
			return -1
		}
		userID = id
	}
	defer rows.Close()
	return userID
}

// AddUser добавляет нового юзера в БД
// Возвращает new User ID
// -1 если была error
func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID != -1 {
		fmt.Println("User already exists:", Username)
		return -1
	}

	insertStatement := `insert into "users" ("username") values ($1)`
	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	userID = exists(d.Username)
	if userID == -1 {
		return userID
	}

	insertStatement = `insert into "userdata" ("userid", "name", "surname", "description")
	values ($1, $2, $3, $4)`
	_, err = db.Exec(insertStatement, userID, d.Name, d.Surname, d.Description)
	if err != nil {
		fmt.Println("db.Exec()", err)
		return -1
	}

	return userID
}

// DeleteUser удаляет существующего юзера
// Запрашивает user ID чтобы удалить
func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Does the ID exist?
	statement := fmt.Sprintf(`SELECT "username" FROM "users" where id = %d`, id)
	rows, err := db.Query(statement)

	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	if exists(username) != id {
		return fmt.Errorf("User with ID %d does not exist", id)
	}

	// Delete from Userdata
	deleteStatement := `delete from "userdata" where userid=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	// Delete from Users
	deleteStatement = `delete from "users" where id=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers получает список всех пользователей в БД и возвращает срез Userdata
func ListUsers() ([]Userdata, error) {
	Data := []Userdata{}
	db, err := openConnection()
	if err != nil {
		return Data, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT "id","username","name","surname","description"
		FROM "users","userdata"
		WHERE users.id = userdata.userid`)
	if err != nil {
		return Data, err
	}

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var description string
		err = rows.Scan(&id, &username, &name, &surname, &description)
		temp := Userdata{ID: id, Username: username, Name: name, Surname: surname, Description: description}
		Data = append(Data, temp)
		if err != nil {
			return Data, err
		}
	}
	defer rows.Close()
	return Data, nil
}

// UpdateUser для обновления существующего пользователя, заданного структурой Userdata
// ID пользователя для обновления находится внутри ф-ции
func UpdateUser(d Userdata) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID == -1 {
		return errors.New("User does not exist")
	}
	d.ID = userID
	updateStatement := `update "userdata" set "name"=$1, "surname"=$2, "description"=$3 where "userid"=$4`
	_, err = db.Exec(updateStatement, d.Name, d.Surname, d.Description, d.ID)
	if err != nil {
		return err
	}

	return nil
}
