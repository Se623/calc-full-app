package lib

import (
	"github.com/Se623/calc-full-app/internal/proto" // Путь к сгенерированным файлам Protobuf
	"google.golang.org/grpc"
)

// Константы
var TIME_ADDITION_MS = 1000        // Время выполнения сложения (мс)
var TIME_SUBTRACTION_MS = 1000     // Время выполнения вычитания (мс)
var TIME_MULTIPLICATIONS_MS = 1000 // Время выполнения умножения (мс)
var TIME_DIVISIONS_MS = 1000       // Время выполнения сложения (мс)
var TIME_REQUESTING_MS = 5000      // Время выполнения сложения (мс)
var COMPUTING_AGENTS = 2           // Кол-во агентов
var COMPUTING_POWER = 2            // Кол-во горутин агентов

// Логгер
var Sugar *zap.SugaredLogger

// Запустить логгер
func InitLogger() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	Sugar = logger.Sugar()
}

// "Сырое" выражение (Инфиксное, не разделённое)
type Raw struct {
	Expression string `json:"expression"`
}

// Выражение
type Expr struct {
	ID       int     `json:"id"`     // Номер выражения
	UserID   int     `json:"userid"` // Номер пользователя выражения
	Oper     string  // Само выражение
	LastTask int     `json:"lasttask"` // Номер последней задачи
	Ans      float64 // Ответ
	Status   int8    // Статус действия: 0 - не решено, 1 - решается, 2 - решено.
	Agent    int
}

// Выражение, которое отображается в API
type ExprDsp struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

// Массив выражений, которые отображаются в API
type DspArr struct {
	Expressions []ExprDsp `json:"expressions"`
}

// Задача
type Task struct {
	ID             int // ID
	ProbID         int // Номер выражения действия
	Link1          int
	Link2          int
	Arg1           float64 // Первое число
	Arg2           float64 // Второе число
	Operation      string  // Операция
	Operation_time int     // Время выполнения
	Ans            float64 // Ответ
	Status         int8    // Статус действия: 0 - не решено, 1 - решается, 2 - решено.
}

// Ответ на задачу по ID
type TaskInc struct {
	ID     int     `json:"id"`
	Result float64 `json:"float64"`
}

type SendExprRequest struct {
	proto.UnimplementedSendExprRequest
}

func (s *userSendExprRequest) GetExpr(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	cand, err := database.DBM.GetNsolEx()
	if err != nil {
		return nil, err
	}
	// Пример данных
	user := &proto.User{
		ID:       cand.ID,
		UserID:   cand.UserID,
		Oper:     cand.Oper,
		LastTask: cand.LastTask,
		Ans:      cand.Ans,
		Status:   cand.Status,
		Agent:    cand.Agent,
	}
	return &proto.UserResponse{User: user}, nil
}

// Стэк
type Stack struct {
	stack []string
}

// Создаёт экземпляр стэка
func Newstack() *Stack {
	return &Stack{stack: []string{}}
}

// Добавляет элемент в стак
func (s *Stack) Push(val string) {
	s.stack = append(s.stack, val)
}

// Просматривает последний элемент в стэке
func (s *Stack) GetTop() string {
	if len(s.stack) != 0 {
		return s.stack[len(s.stack)-1]
	} else {
		return ""
	}
}

// Вынимает последний элемент из стэка
func (s *Stack) Pop() string {
	if len(s.stack) != 0 {
		r := s.stack[len(s.stack)-1]
		s.stack = s.stack[:len(s.stack)-1]
		return r
	} else {
		return ""
	}
}
