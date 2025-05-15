package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Se623/calc-full-app/internal/database"
	"github.com/Se623/calc-full-app/internal/lib"
	"github.com/Se623/calc-full-app/pkg/rpn"
)

func Displayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		exprArr, err := database.DBM.GetAllExpr()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		exprArrPack, err := json.Marshal(exprArr)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(exprArrPack))

	} else {
		var exprPack []byte
		var err error
		idInt, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Error: ID not found", http.StatusNotFound)
			return
		}
		cand, err := database.DBM.GetExpr(idInt)
		if err != nil {
			if err.Error() == "no expressions" {
				http.Error(w, "Error: Expression not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if cand.Status == 0 {
			exprPack, err = json.Marshal(lib.ExprDsp{ID: cand.ID, Status: "Queued", Result: -1})
		} else if cand.Status == 1 {
			exprPack, err = json.Marshal(lib.ExprDsp{ID: cand.ID, Status: "Solving", Result: -1})
		} else if cand.Status == 2 {
			exprPack, err = json.Marshal(lib.ExprDsp{ID: cand.ID, Status: "Solved", Result: cand.Ans})
		}

		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `{"expression": "%s"}`, string(exprPack))
	}
}

func Spliter(w http.ResponseWriter, r *http.Request) {
	pr := []string{}
	res := [][]string{}
	linkctr := 0

	rpnstack := lib.Newstack()

	decoder := json.NewDecoder(r.Body)
	var resp lib.Raw
	err := decoder.Decode(&resp)

	if err != nil {
		http.Error(w, "Error: Invalid JSON", http.StatusInternalServerError)
		return
	}

	rpnarr, err := rpn.InfixToPostfix(resp.Expression)
	if err != nil {
		http.Error(w, "Error: Invalid Input", http.StatusUnprocessableEntity)
		return
	}

	// Разделение выражений на задания
	for _, v := range rpnarr {
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			rpnstack.Push(v)
		} else {
			rawo2 := rpnstack.Pop()
			rawo1 := rpnstack.Pop()
			if rawo1 == "" || rawo2 == "" {
				http.Error(w, "Error: Invalid Input", http.StatusUnprocessableEntity)
				return
			}
			pr = append(pr, rawo2)
			pr = append(pr, rawo1)

			pr = append(pr, v)
			rpnstack.Push("L-" + fmt.Sprint(linkctr))
			linkctr++

			res = append(res, pr)
			pr = []string{}
		}
	}

	nums := []int{}

	for _, v := range res {
		v[0], v[1] = v[1], v[0]

		var a float64
		var b float64

		optime := 0
		links := [2]int{-1, -1}

		if v[2] == "+" {
			optime = lib.TIME_ADDITION_MS
		} else if v[2] == "-" {
			optime = lib.TIME_SUBTRACTION_MS
		} else if v[2] == "*" {
			optime = lib.TIME_MULTIPLICATIONS_MS
		} else if v[2] == "/" {
			optime = lib.TIME_DIVISIONS_MS
		}

		// 76 - числовое значение L
		if v[0][0] == 76 {
			links[0], _ = strconv.Atoi(v[0][2:])
			a = -1
		} else {
			a, _ = strconv.ParseFloat(v[0], 64)
		}
		if v[1][0] == 76 {
			links[1], _ = strconv.Atoi(v[1][2:])
			b = -1
		} else {
			b, _ = strconv.ParseFloat(v[1], 64)
		}

		if len(nums) != 0 {
			if links[0] != -1 {
				links[0] = links[0] + nums[0]
			}
			if links[1] != -1 {
				links[1] = links[1] + nums[0]
			}
		}

		taskid, err := database.DBM.AddTask(lib.Task{ID: -1, ProbID: 0, Link1: links[0], Link2: links[1], Arg1: a, Arg2: b, Operation: v[2], Operation_time: optime, Ans: 0, Status: 0})
		if err != nil {
			lib.Sugar.Errorf("Orchestrator: Got error when spliting: %s", err.Error())
		}
		nums = append(nums, taskid)
		fmt.Println(nums)
	}

	exprid, err := database.DBM.AddExpr(lib.Expr{ID: 0, UserID: 0, Oper: resp.Expression, LastTask: nums[len(nums)-1], Ans: 0, Status: 0, Agent: -1})
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	for _, v := range nums {
		database.DBM.UpdTaskID(v, exprid)
	}
	fmt.Fprintf(w, `{"id": "%d"}`, exprid)
}

func Distributor() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		lib.Sugar.Fatalf("Error launching distributor: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &userServiceServer{})

	lib.Sugar.Infof("Distributer launched on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		lib.Sugar.Fatalf("Error launching distributor: %v", err)
	}
}
