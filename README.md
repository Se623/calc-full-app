# calc-full-app

Это распределённый калькулятор на golang. Он роасполагается на http сервере. Что-бы им воспользоваться, нужно получить токен по запросу 'localhost:8080/api/v1/register' с JSON типа '{ "login": , "password": }. Если нет логина и пароля, зарегистрироватся можно по адресу 'localhost:8080/api/v1/register' с JSON типа '{ "login": , "password": } (Нельзя делать аккаунты с одинаковыми аккаунтами и паролями!). Токен перестаёт быть действительным через день после его выдачи. После получения токена можно послать POST запрос на 'localhost:8080/api/v1/calculate' С JSON типа '{"expression": "выражение"}', и он пришлёт ответ с JSON типа {"id": "ID выражения"} и код 201. Это означает, что выражение находится в очереди на решение. Он также может выдать 2 кода ошибок:

    422 - Если выражение не соответствуют требованиям приложения (Нелегальные символы, или выражение не решаемо.)
    500 - Если в теле запроса есть ошибки (Запрос не оформлен по правилам JSON)

Чтобы просмотреть выражения, которые находятся у сервера, нужно послать GET запрос на 'localhost/api/v1/expressions', сервер пришлёт список со выражениями, корорые вы посылали из вашего аккаунта(токена). Элемент этого списка - JSON с тремя полями:
+ id - id выражения, которое даётся при загрузке этого выражения в систему
+ status - статус выражения, может быть "Queued" - выражение ожидает свободного агента, "Solving" - выражение решается агентом, "Solved" - выражение решено
+ result - результат выражения, при статусах "Queued" и "Solving" - result всегда -1


## Запуск

1. Cклонировать репозиторий (Нужна программа git)
```bash
git clone https://github.com/Se623/calc-full-app
```
2. Перейти в директорию программы
```bash
cd ./calc-full-app
```
3. Установить зависимости
```bash
go install ./cmd
```
4. Запустить калькулятор
```bash
go run ./cmd
```

Сервер распологается на порту 8080.

## Примеры

### Регистрация
Запрос:\
Bash(Linux): `curl --location 'localhost:8080/api/v1/register' --header 'Content-Type: application/json' --data '{ "login":"asd" , "password":"asd"}'`\
Cmd: `curl --location "localhost:8080/api/v1/register" --header "Content-Type: application/json" --data '{ \"login\":\"asd\" , \"password\":\"asd\"}"`\

Ответ: `{"id": "0"}` 

### Логин
Запрос:\
Bash(Linux): `curl --location 'localhost:8080/api/v1/login' --header 'Content-Type: application/json' --data '{ "login":"asd" , "password":"asd"}''`\
Cmd: `curl --location 'localhost:8080/api/v1/login' --header "Content-Type: application/json" --data '{ \"login\":\"asd\" , \"password\":\"asd\"}'"`\

Ответ: `{"token": "<Очень длинный токен>"}` 


### Пример 1 (Обычное выражение)
Запрос:\
Bash(Linux): `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' --data '{"expression": "2+2*2"}'`\
Cmd: `curl --location "localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --header "Authorization: Bearer <токен>" --data "{\"expression\": \"2+2*2\"}"`\

Ответ: `{"id": "0"}` (Ответ на выражение: 6)

### Пример 2 (Сложное выражение)
Запрос:\
Bash(Linux): `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' --data '{"expression": "(6+8.2)*5.12-(5.971-8.3335)/5"}'`\
Cmd: `curl --location "localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --header "Authorization: Bearer <токен>" --data "{\"expression\": \"(6+8.2)*5.12-(5.971-8.3335)/5\"}"`

Ответ: `{"id": "1"}` (Ответ на выражение: 73.17649999999999)

### Пример 3 (Ошибка)
Запрос:\
Bash(Linux): `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' --data '{"expression": "***5***"}'`\
Cmd: `curl --location "localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --header "Authorization: Bearer <токен>" --data "{\"expression\": \"***5***\"}"`\

Ответ: `Error: Invalid Input` (Выражение не покажется в списке)

### Пример 4 (Ошибка c JSON)
Запрос:\
Bash(Linux): `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --header 'Authorization: Bearer <токен>' --data 'asdfg'`\
Cmd: `curl --location "localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --header "Authorization: Bearer <токен>" --data "asdfg"`\

Ответ: `Error: Invalid JSON` (Выражение не покажется в списке)

## Пример просмотра выражений
### Запрос всех выражений
Запрос:\
Bash(Linux): `curl --location localhost:8080/api/v1/expressions --header 'Authorization: Bearer <токен>'`
Cmd: `curl --location localhost:8080/api/v1/expressions --header "Authorization: Bearer <токен>"`

Ответ: `{"expressions":[{"id":0,"status":"Queued","result":-1},{"id":1,"status":"Solved","result":6}]}` (Может быть другой ответ в зависимости от решаемых выражений)

### Запрос выражения по id
Запрос:\
Bash(Linux): `curl --location localhost:8080/api/v1/expressions?id=1 --header 'Authorization: Bearer <токен>'`
Cmd: `curl --location localhost:8080/api/v1/expressions?id=1 --header "Authorization: Bearer <токен>"`


Ответ: `{"expressions":[{"id":0,"status":"Queued","result":-1}]}` (Может быть другой ответ в зависимости от решаемых выражений)





