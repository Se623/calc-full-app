package database

import (
	"database/sql"
	"errors"
	"sync"

	lib "github.com/Se623/calc-full-app/internal/lib"
	_ "github.com/mattn/go-sqlite3"
)

type DBmutex struct {
	mutex sync.Mutex
}

var DBM DBmutex

// Создаёт базу данных (Обновляет если существует)
func Init() error {
	DBM = DBmutex{}
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return err
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS expressions (
		id				INTEGER PRIMARY KEY AUTOINCREMENT,
		userid			INTEGER,
		oper  			VARCHAR(255),
		lasttask		INTEGER,
		ans				REAL,
		status			TINYINT,
		agent			INTEGER
	  );
	CREATE TABLE IF NOT EXISTS tasks (
		id        		INTEGER PRIMARY KEY AUTOINCREMENT,
		probid			INTEGER NOT NULL,
		link1 			INTEGER,
		link2			INTEGER,
		arg1          	REAL,
		arg2           	REAL,
		operation      	VARCHAR(255),
		operation_time 	INTEGER, 
		ans            	REAL,
		status         	TINYINT  
	  );
	CREATE TABLE IF NOT EXISTS users (
		id        		INTEGER PRIMARY KEY AUTOINCREMENT,
		login			VARCHAR(255),
		password		VARCHAR(255)
	  );
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// Возращает пользователя по паролю
func (d *DBmutex) AddUser(user string, pass string) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return 0, err
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO users (login, password) VALUES (?, ?);", user, pass)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Возращает пользователя по паролю
func (d *DBmutex) CheckUser(user string, pass string) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT ID FROM users WHERE login = ? AND password = ?", user, pass)

	var p int

	if err := row.Scan(&p); err != nil {
		return 0, err
	}
	return p, nil
}

// Добавляет выражение в базу данных (Если это выражение уже существует - функция обновляет его)
func (d *DBmutex) AddExpr(expr lib.Expr) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return 0, err
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO expressions (userid, oper, lasttask, ans, status, agent) VALUES (?, ?, ?, ?, ?, ?);", expr.UserID, expr.Oper, expr.LastTask, expr.Ans, expr.Status, expr.Agent)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (d *DBmutex) AddTask(task lib.Task) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return 0, err
	}

	defer db.Close()

	err = db.QueryRow("SELECT id FROM tasks WHERE id = $1;", task.ID).Scan(&task.ID)

	if err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}

		result, err := db.Exec("INSERT INTO tasks (probid, link1, link2, arg1, arg2, operation, operation_time, ans, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", task.ProbID, task.Link1, task.Link2, task.Arg1, task.Arg2, task.Operation, task.Operation_time, task.Ans, task.Status)

		if err != nil {
			return 0, err
		}

		id, err := result.LastInsertId()

		if err != nil {
			return 0, err
		}

		return int(id), nil
	}

	_, err = db.Exec("UPDATE tasks SET probid = ?, link1 = ?, link2 = ?, arg1 = ?, arg2 = ?, operation = ?, operation_time = ?, ans = ?, status = ? WHERE id = ?", task.ProbID, task.Link1, task.Link2, task.Arg1, task.Arg2, task.Operation, task.Operation_time, task.Ans, task.Status, task.ID)

	if err != nil {
		return 0, err
	}

	return -1, nil
}

// Обновляет ответ, статус и агента выражения
func (d *DBmutex) UpdExpr(id int, status int8, agent int, ans float64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE expressions SET ans = ?, status = ?, agent = ? WHERE id = ?;", ans, status, agent, id)
	if err != nil {
		return err
	}
	return nil
}

// Обновляет ответ и статус задачи
func (d *DBmutex) UpdTask(id int, status int8, ans float64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE tasks SET ans = ?, status = ? WHERE id = ?;", ans, status, id)
	if err != nil {
		return err
	}
	return nil
}

// Обновляет ответ и статус задачи
func (d *DBmutex) UpdAllTaskSt(probid int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE tasks SET status = 0 WHERE probid = ?;", probid)
	if err != nil {
		return err
	}
	return nil
}

// Обновляет ответ и статус задачи
func (d *DBmutex) UpdTaskID(id int, probid int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE tasks SET probid = ? WHERE id = ?;", probid, id)
	if err != nil {
		return err
	}
	return nil
}

// Удаляет операцию из базы данных
func (d *DBmutex) DelTask(id int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?;", id)
	if err != nil {
		return err
	}
	return nil
}

// Возращает массив из выражений
func (d *DBmutex) GetAllExpr(userid int) (lib.DspArr, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.DspArr{}, err
	}
	defer db.Close()

	var exprs lib.DspArr

	rows, err := db.Query("SELECT id, status, ans FROM expressions WHERE userid = ?", userid)
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
func (d *DBmutex) GetExpr(id int) (lib.Expr, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE ID = ?", id)

	var p lib.Expr

	if err := row.Scan(&p.ID, &p.UserID, &p.Oper, &p.LastTask, &p.Ans, &p.Status, &p.Agent); err != nil {
		return lib.Expr{}, err
	}
	return p, nil
}

// Достаёт выражение по агенту
func (d *DBmutex) GetExprAg(agent int) (lib.Expr, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE agent = ?", agent)

	var p lib.Expr

	if err := row.Scan(&p.ID, &p.UserID, &p.Oper, &p.LastTask, &p.Ans, &p.Status, &p.Agent); err != nil {
		return lib.Expr{}, err
	}
	return p, nil
}

// Достаёт операцию по ID
func (d *DBmutex) GetTask(exprid int, id int) (lib.Task, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.Task{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM tasks WHERE ID = ? AND probid = ?", id, exprid)

	var p lib.Task

	if err := row.Scan(&p.ID, &p.ProbID, &p.Link1, &p.Link2, &p.Arg1, &p.Arg2, &p.Operation, &p.Operation_time, &p.Ans, &p.Status); err != nil {
		return lib.Task{}, err
	}
	return p, nil
}

// Достаёт операцию по ссылкам
func (d *DBmutex) GetTaskLk(exprid int, lknm int, link int) (lib.Task, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")
	if err != nil {
		return lib.Task{}, err
	}
	defer db.Close()

	var row *sql.Row

	if lknm == 1 {
		row = db.QueryRow("SELECT * FROM tasks WHERE probid = ? AND link1 = ?", exprid, link)
	} else if lknm == 2 {
		row = db.QueryRow("SELECT * FROM tasks WHERE probid = ? AND link2 = ?", exprid, link)
	} else {
		return lib.Task{}, errors.New("unknown link")
	}

	var p lib.Task

	if err := row.Scan(&p.ID, &p.ProbID, &p.Link1, &p.Link2, &p.Arg1, &p.Arg2, &p.Operation, &p.Operation_time, &p.Ans, &p.Status); err != nil {
		return lib.Task{}, err
	}
	return p, nil
}

// Возращает первое нерешённое выражение
func (d *DBmutex) GetNsolEx() (lib.Expr, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return lib.Expr{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM expressions WHERE status = 0 AND agent = -1")

	var p lib.Expr
	if err := row.Scan(&p.ID, &p.UserID, &p.Oper, &p.LastTask, &p.Ans, &p.Status, &p.Agent); err != nil {
		if err != sql.ErrNoRows {
			return lib.Expr{}, err
		}
		return lib.Expr{}, errors.New("no expressions")
	}
	return p, nil
}

// Возращает первую нерешённую задачу выражения
func (d *DBmutex) GetNsolTs(id int) (lib.Task, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, err := sql.Open("sqlite3", "./calc.db")

	if err != nil {
		return lib.Task{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM tasks WHERE status = 0 AND link1 = -1 AND link2 = -1")

	var p lib.Task
	if err := row.Scan(&p.ID, &p.ProbID, &p.Link1, &p.Link2, &p.Arg1, &p.Arg2, &p.Operation, &p.Operation_time, &p.Ans, &p.Status); err != nil {
		return lib.Task{}, err
	}
	return p, nil
}
