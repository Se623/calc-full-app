package database

import (
	"database/sql"
	"errors"

	lib "github.com/Se623/calc-full-app/internal/lib"
	_ "github.com/mattn/go-sqlite3"
)

// Создаёт базу данных (Обновляет если существует)
func Init() error {

	db, err := sql.Open("sqlite3", "./calc.db")

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
		link1 			INT,
		link2			INT,
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
func AddExpr(expr lib.Expr) (int, error) {
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
	return int(id), nil
}

func AddTask(task lib.Task) (int, error) {
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return 0, err
	}

	defer db.Close()

	err = db.QueryRow("SELECT id FROM operation WHERE id = $1;", task.ID).Scan(&task.ID)

	if err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}

		result, err := db.Exec("INSERT INTO expressions (probid, link1, link2, Arg1, Arg2, Operation, Operation_time, Status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", task.ProbID, task.Link1, task.Link2, task.Arg1, task.Arg2, task.Operation, task.Operation_time, task.Status)

		if err != nil {
			return 0, err
		}

		id, err := result.LastInsertId()

		if err != nil {
			return 0, err
		}

		return int(id), nil
	}

	_, err = db.Exec("UPDATE operation SET probid = ?, link1 = ?, link2 = ?, Arg1 = ?, Arg2 = ?, Operation = ?, Operation_time = ?, Status = ?", task.ProbID, task.Link1, task.Link2, task.Arg1, task.Arg2, task.Operation, task.Operation_time, task.Status)

	if err != nil {
		return 0, err
	}

	return -1, nil
}

// Обновляет ответ и статус выражения
func UpdExpr(id int, status int8, ans float64) error {
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

// Обновляет ответ и статус задачи
func UpdTask(id int, status int8, ans float64) error {
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE task SET ans = ?, status = ? WHERE id = ?;", ans, status, id)
	if err != nil {
		return err
	}
	return nil
}

// Удаляет операцию из базы данных
func DelTask(id int) error {
	db, err := sql.Open("sqlite3", "./calc.db")
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

// Возращает массив из выражений
func GetAllExpr() (lib.DspArr, error) {
	db, err := sql.Open("sqlite3", "./calc.db")
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

// Достаёт выражение по ID
func GetExpr(id int) (lib.Expr, error) {
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE ID = ?", id)

	var p lib.Expr

	if err := row.Scan(&p.ID, &p.Oper, &p.LastTask, &p.Ans, &p.Status); err != nil {
		return lib.Expr{}, err
	}
	return p, nil
}

// Достаёт операцию по ID и ссылкам
func GetTask(exprid int, id int) (lib.Task, error) {
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.Task{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM tasks WHERE ID = ?", id)

	var p lib.Task

	if err := row.Scan(&p.ID, &p.ProbID, &p.Link1, &p.Link2, &p.Arg1, &p.Arg2, &p.Operation, &p.Operation_time, &p.Ans, &p.Status); err != nil {
		return lib.Task{}, err
	}
	return p, nil
}

// Возращает первое нерешённое выражение
func GetNsolEx() (lib.Expr, error) {
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE solving = 0")

	var p lib.Expr
	if err := row.Scan(&p.ID, &p.Oper, &p.LastTask, &p.Ans, &p.Status); err != nil {
		if err != sql.ErrNoRows {
			return lib.Expr{}, err
		}
		return lib.Expr{}, errors.New("no expressions")
	}
	return p, nil
}

// Возращает первую нерешённую задачу выражения
func GetNsolTs(id int) (lib.Task, error) {
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return lib.Task{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE status = 0 AND link1 = 0 AND link2 = 0")

	var p lib.Task
	if err := row.Scan(&p.ID, &p.ProbID, &p.Link1, &p.Link2, &p.Arg1, &p.Arg2, &p.Operation, &p.Operation_time, &p.Ans, &p.Status); err != nil {
		if err != sql.ErrNoRows {
			return lib.Task{}, err
		}
		return lib.Task{}, errors.New("no expressions")
	}
	return p, nil
}
