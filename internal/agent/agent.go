package agent

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Se623/calc-full-app/internal/database"
	"github.com/Se623/calc-full-app/internal/lib"
)

func Agent(id int) {
	lib.Sugar.Infof("Agent %d: I am started", id)
	ticker := time.NewTicker(time.Duration(lib.TIME_REQUESTING_MS) * time.Millisecond)

	calcsin := make(chan lib.Task)
	calcsout := make(chan lib.Task)
	exprslot := lib.Expr{}
	busy := false

	cand, err := database.DBM.GetExprAg(id)

	if err == nil {
		exprslot = cand
		busy = true
		lib.Sugar.Infof("Agent %d: Recovered task %d", id, cand.ID)
		database.DBM.UpdAllTaskSt(cand.ID)
	}

	for i := 0; i < lib.COMPUTING_POWER; i++ {
		go Calculator(calcsin, calcsout, i+1, id)
	}

	for {
		select {
		case resTask := <-calcsout:
			if resTask.Status == 2 {
				lib.Sugar.Infof("Agent %d: Got task %d", id, resTask.ID)

				err := database.DBM.UpdTask(resTask.ID, resTask.Status, resTask.Ans)
				fmt.Println(err)
				cand1, err := database.DBM.GetTaskLk(exprslot.ID, 1, resTask.ID)
				fmt.Println(err, resTask.ID, exprslot.ID)
				if err == nil {
					fmt.Println(cand1)
					cand1.Arg1 = resTask.Ans
					cand1.Link1 = -1
					_, _ = database.DBM.AddTask(cand1)
					lib.Sugar.Infof("Agent %d: Changed 1st link in task %d to number", id, cand1.ID)
				}
				cand2, err := database.DBM.GetTaskLk(exprslot.ID, 2, resTask.ID)
				if err == nil {
					cand2.Arg2 = resTask.Ans
					cand2.Link2 = -1
					_, _ = database.DBM.AddTask(cand2)
					lib.Sugar.Infof("Agent %d: Changed 2nd link in task %d to number", id, cand2.ID)
				}

				if resTask.ID == exprslot.LastTask {
					lib.Sugar.Infof("Agent %d: Task %d - last task, sending to orchestartor", id, resTask.ID)
					data, _ := json.Marshal(lib.TaskInc{ID: exprslot.ID, Result: resTask.Ans})
					r := bytes.NewReader(data)
					_, err := http.Post("http://localhost:8080/internal/task", "application/json", r)
					if err != nil {
						fmt.Println(err)
					}
					busy = false
				}
				database.DBM.DelTask(resTask.ID)
				continue

			}
		case <-ticker.C:
			if !busy {
				conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
				if err != nil {
					log.Fatalf("Error: %v", err)
				}
				defer conn.Close()

				client := pb.NewUserServiceClient(conn)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				req := &pb.UserRequest{}
				expr, err := client.GetUser(ctx, req)
				if err != nil {
					log.Fatalf("Error: %v", err)
				}

				cand, _ := database.DBM.GetExpr(expr.ID)
				if cand.Agent != -1 {
					lib.Sugar.Errorf("Agent %d: Expression blocked", id)
					continue
				}
				database.DBM.UpdExpr(expr.ID, 1, id, 0)
				lib.Sugar.Infof("Agent %d: Got expression %d", id, expr.ID)
				exprslot = expr
				busy = true
			}
		default:
			if busy {
				cand, err := database.DBM.GetNsolTs(exprslot.ID)
				if err != nil {
					if err != sql.ErrNoRows {
						lib.Sugar.Errorf("Agent %d: Got an error during tasks search: %s", id, err.Error())
					}
					break
				} else {
					lib.Sugar.Infof("Agent %d: Got undestributed task %d, sending to calculators", id, cand.ID)
					calcsin <- cand
					database.DBM.UpdTask(cand.ID, 1, -1)
				}
			}
		}
	}
}

func Calculator(comm chan lib.Task, result chan lib.Task, id int, agid int) {
	lib.Sugar.Infof("Calc %d-%d: I am started", agid, id)
	for task := range comm {
		lib.Sugar.Infof("Calc %d-%d: Got task %d: %d %s %d", agid, id, task.ID, task.Arg1, task.Operation, task.Arg2)
		timer := time.NewTimer(time.Duration(task.Operation_time) * time.Millisecond)
		<-timer.C
		if task.Operation == "+" {
			task.Ans = task.Arg1 + task.Arg2
		} else if task.Operation == "-" {
			task.Ans = task.Arg1 - task.Arg2
		} else if task.Operation == "*" {
			task.Ans = task.Arg1 * task.Arg2
		} else if task.Operation == "/" {
			task.Ans = task.Arg1 / task.Arg2
		}
		lib.Sugar.Infof("Calc %d-%d: Got answer to the task %d: %d", agid, id, task.ID, task.Ans)
		task.Status = 2
		result <- task
	}
}
