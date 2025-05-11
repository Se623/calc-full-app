package main

import (
	"net/http"

	"github.com/Se623/calc-full-app/internal/agent"
	"github.com/Se623/calc-full-app/internal/database"
	"github.com/Se623/calc-full-app/internal/lib"
	"github.com/Se623/calc-full-app/internal/orchestrator"
)

func main() {
	lib.InitLogger()

	lib.Sugar.Info("Initilized main program")

	for i := 0; i < lib.COMPUTING_AGENTS; i++ {
		go agent.Agent(i + 1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", orchestrator.Spliter)
	mux.HandleFunc("/internal/task", orchestrator.Distributor)
	mux.HandleFunc("localhost/api/v1/expressions", orchestrator.Displayer)
	lib.Sugar.Infof("Initilized orchestrator")

	err := database.Init()
	if err != nil {
		lib.Sugar.Errorf(err.Error())
		return
	}
	lib.Sugar.Infof("Initilized database")

	if err = http.ListenAndServe(":8080", mux); err != nil { // Запуск сервера
		lib.Sugar.Errorf(err.Error())
		return
	}
	lib.Sugar.Infof("Initilized server on port 8080")
}
