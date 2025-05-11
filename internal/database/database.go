package database

import (
	"database/sql"
	"errors"

	lib "github.com/Se623/calc-full-app/internal/lib"
	_ "github.com/mattn/go-sqlite3"
)

// Создаёт базу данных (Обновляет если существует)
func Init() error {

	db, err := sql.Open("sqlite3", "./expressions.db")

	if err != nil {
		return err
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS expressions (
		id				INTEGER PRIMARY KEY AUTOINCREMENT,
		userid			INTEGER
		oper  			VARCHAR(255),
		tasksid			VARCHAR(255),
		ans				REAL,
		status			TINYINT
	  );
	CREATE TABLE IF NOT EXISTS tasks (
		id        		INTEGER PRIMARY KEY AUTOINCREMENT,
		probid			INT NOT NULL,
		links 			VARCHAR(255),
		Arg1          	REAL,
		Arg2           	REAL,
		Operation      	VARCHAR(255),
		Operation_time 	INT, 
		Ans            	REAL,
		Status         	TINYINT  
	  );
	CREATE TABLE IF NOT EXISTS users (
		id        		INTEGER PRIMARY KEY AUTOINCREMENT,
		password		VARCHAR(255)
	  );
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

// Добавляет выражение в базу данных (Если это выражение уже существует - функция обновляет его)
func AddExpr(expr lib.Expr) (int64, error) {
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return 0, err
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO expressions (oper, tasksid) VALUES (?, ?);", expr.Oper, "a s d")
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Возращает массив из выражений
func GetAllExpr() (lib.DspArr, error) {
	db, err := sql.Open("sqlite3", "./expressions.db")
	if err != nil {
		return lib.DspArr{}, err
	}
	defer db.Close()

	var exprs lib.DspArr

	rows, err := db.Query("SELECT id, status, ans FROM problem")
	if err != nil {
		return lib.DspArr{}, err
	}
	for rows.Next() {
		var p lib.ExprDsp
		if err := rows.Scan(&p.ID, &p.Status, &p.Result); err != nil {
			return lib.DspArr{}, err
		}
		if p.Status == "0" {
			p.Status = "Queued"
		} else if p.Status == "1" {
			p.Status = "Solving"
		} else if p.Status == "2" {
			p.Status = "Solved"
		}
		exprs.Expressions = append(exprs.Expressions, p)
	}
	if err := rows.Err(); err != nil {
		return lib.DspArr{}, err
	}
	return exprs, nil
}

// Удаляет все операции выражения из базы данных
func DelProbChilds(id int64) error {
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM operation WHERE probid = ?;", id)
	if err != nil {
		return err
	}

	return nil
}

// Обновляет ответ и статус выражения
func UpdExpr(id int64, status int, ans float64) error {
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE expressions SET ans = ?, status = ? WHERE id = ?;", ans, status, id)
	if err != nil {
		return err
	}
	return nil
}

// Добавляет операцию в базу данных (Если это выражение уже существует - функция обновляет его)
func AddOper(oper lib.Expr) error {
	db, err := sql.Open("sqlite3", "./expressions.db")
	s := 0
	l := 0

	if err != nil {
		return err
	}
	defer db.Close()

	if oper.Solving {
		s = 1
	}
	if oper.Last {
		l = 1
	}

	err = db.QueryRow("SELECT id FROM operation WHERE id = $1;", oper.ID).Scan(&oper.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		_, err = db.Exec("INSERT INTO operation (probid, operation, ans, solving, last) VALUES (?, ?, ?, ?, ?)", oper.ProbID, oper.Expr, oper.Ans, s, l)
		if err != nil {
			return err
		}
		return nil
	}
	_, err = db.Exec("UPDATE operation SET probid = ?, operation = ?, ans = ?, solving = ?, last = ? WHERE id = ?;", oper.ProbID, oper.Expr, oper.Ans, s, l, oper.ID)
	if err != nil {
		return err
	}
	return nil
}

// Удаляет операцию из базы данных
func DelOper(id int64) error {
	db, err := sql.Open("sqlite3", "./expressions.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM operation WHERE id = ?;", id)
	if err != nil {
		return err
	}
	return nil
}

// Достаёт выражение по ID
func GetExpr(id int64) (lib.Expr, error) {
	db, err := sql.Open("sqlite3", "./expressions.db")
	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expression WHERE ID = ?", id)

	var p lib.Expr

	if err := row.Scan(&p.ID, &p.ProbID, &p.Expr, &p.Ans, &p.Solving, &p.Last); err != nil {
		return lib.Expr{}, err
	}
	return p, nil
}

// Достаёт операцию по ID
func GetTask(id int64) (lib.Task, error) {
	db, err := sql.Open("sqlite3", "./expressions.db")
	if err != nil {
		return lib.Task{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM operation WHERE ID = ?", id)

	var p lib.Task

	if err := row.Scan(&p.ID, &p.ProbID, &p.Expr, &p.Ans, &p.Solving, &p.Last); err != nil {
		return lib.Task{}, err
	}
	return p, nil
}

// Возращает первое нерешённое выражение
func GetNsolEx() (lib.Expr, error) {
	db, err := sql.Open("sqlite3", "./expressions.db")

	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE solving = 0")

	err = row.Scan()
	if err != nil {
		if err != sql.ErrNoRows {
			return lib.Expr{}, err
		}
		return lib.Expr{}, errors.New("no expressions")
	}

	var p lib.Expr
	if err := row.Scan(&p.ID, &p.ProbID, &p.Expr, &p.Ans, &p.Solving, &p.Last); err != nil {
		return lib.Expr{}, err
	}
	return p, nil
}
