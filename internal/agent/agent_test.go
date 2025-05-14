package agent_test

import (
	"testing"
	"time"

	"github.com/Se623/calc-full-app/internal/agent"
	"github.com/Se623/calc-full-app/internal/lib"
)

func TestCalculator(t *testing.T) {

	comm := make(chan lib.Task)
	result := make(chan lib.Task)

	go agent.Calculator(comm, result, 1, 1)

	testCases := []struct {
		name           string
		task           lib.Task
		expectedResult float64
	}{
		{
			name: "Addition",
			task: lib.Task{
				ID:             1,
				Arg1:           5,
				Arg2:           3,
				Operation:      "+",
				Operation_time: 10,
				Status:         1,
			},
			expectedResult: 8,
		},
		{
			name: "Subtraction",
			task: lib.Task{
				ID:             2,
				Arg1:           10,
				Arg2:           4,
				Operation:      "-",
				Operation_time: 10,
				Status:         1,
			},
			expectedResult: 6,
		},
		{
			name: "Multiplication",
			task: lib.Task{
				ID:             3,
				Arg1:           7,
				Arg2:           3,
				Operation:      "*",
				Operation_time: 10,
				Status:         1,
			},
			expectedResult: 21,
		},
		{
			name: "Division",
			task: lib.Task{
				ID:             4,
				Arg1:           20,
				Arg2:           5,
				Operation:      "/",
				Operation_time: 10,
				Status:         1,
			},
			expectedResult: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			comm <- tc.task

			select {
			case res := <-result:
				if res.Ans != tc.expectedResult {
					t.Errorf("Expected result %f, got %f", tc.expectedResult, res.Ans)
				}
				if res.Status != 2 {
					t.Errorf("Expected status 2, got %d", res.Status)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("Timeout waiting for result")
			}
		})
	}

	close(comm)
	close(result)
}
