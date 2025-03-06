# Калькулятор распределенных арифметических выражений

[English](README.md)

Telegram: @neo536

## Описание

Демонстрация распределенного веб-сервера, использующего веб-API через JSON для выполнения арифметических вычислений. Агент — калькулятор для веб-сайта.

## Запуск

### Ленивый запуск

При необходимости используйте `chmod +x lazylaunch.sh`, чтобы добавить права на исполнение.

Для запуска:

```bash
./lazylaunch
```

В директории проекта.

### Ручной запуск

Откройте два терминала и запустите:

* `go run cmd/orchestrator/main.go`
* `go run cmd/agent/main.go`

В этом порядке

### Редактирование env

Переменные env для продолжительности операций и количества потоков хранятся в env.env

### Тестирование

Чтобы отправить выражение, используйте файл `python3 manualapi.py` для легкого доступа или команды `curl` ниже:

#### Отправка выражений

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "(50 * (2 + 3))   / ((9 - 6) * (3)) + 4"}'
```

#### Получение выражений

```bash
curl --location 'localhost:8080/api/v1/expressions'
```

`> {"expressions":[{"id":8,"result":-424,"status":"completed"}`

#### Получение определенного выражения

```bash
curl --location 'localhost:8080/api/v1/expressions/8'
```

`> {"expression":{"id":8,"result":-424,"status":"completed"}}`

## Архитектура системы

```text
* - - - - - - - *                 * - - - - - - - *
|  orchestrator | - tasks (2+2) > |     agent     |
* - - - - - - - *                 * - - - - - - - *
        ^
   expressions
        |
* - - - - - - - *
|      user     |
* - - - - - - - *

```

Проект состоит из двух компонентов:

* Агент — Калькулятор, который выполняет отдельные арифметические задачи, полученные от оркестратора
* Оркестратор — Бэкэнд, который получает вызовы API и вычисляет порядок операций для агента.

### Оркестратор

Это веб-сервер. Он управляет вызовами API и задачами.

При получении нового выражения оно обрабатывается следующим образом:

1. Проверка проверяет выражение, устраняя необходимость в дополнительных проверках (паника все еще сохраняется, когда что-то идет совсем не так)
2. Выражение получает отдельные числа, объединенные заново
3. Выражение токенизируется в [][]float
4. Выражение преобразуется в AST с помощью алгоритма рекурсивного спуска-анализатора, разделяющего выражения, термины и факторы
5. AST преобразуется в список задач с деревом зависимостей

При вызове /internal/task неотправленная, неготовая задача отправляется агенту

При получении результата он помечается как выполненный, и теперь другая задача готова к отправке

*Все идентификаторы выражений являются идентификаторами задач, но только идентификаторы родительских задач являются идентификаторами выражений.* Это упрощает процесс.

В приведенном выше примере задача с `"id": 8` является корнем дерева задач, содержащего задачи с идентификаторами 0-8.

### Задача

Задача имеет следующую структуру:

```golang
type Task struct {
 ID       int      // ID
 Operator string   // Строка оператора (если это значение, то пусто)
 LeftID   *int     // ID левого листа или nil
 RightID  *int     // ID правого листа или nil
 LeftVal  *float64
 RightVal *float64
 parent   bool     // Является ли это корнем дерева
 Done     bool     // Готова ли задача
 Result   *float64
 Time     int64
 Sent     bool     // Есть ли у агента эта задача
}
```

## Конечные точки API клиента

### POST /api/v1/calculate

Request:

```json
{
  "expression": "<expression here>"
}
```

Response:

```json
{
    "id": <unique id of parent task>
}
```

### GET /api/v1/expressions

Response:

```json
{
    "expressions": [
        {
            "id": <unique id of parent task>,
            "status": <status>,
            "result": <result>
        },
        ...
    ]
}
```

### GET /api/v1/expressions/{id}

Response:

```json
{
    "expression":
        {
            "id": <unique id of expr>,
            "status": <status>,
            "result": <result>
        }
}
```

## Agent API Endpoints

### GET /internal/task

```json
{
    "task":
        {
            "id": <unique task id>,
            "arg1": <first arg>,
            "arg2": <second arg>,
            "operation": <operation>,
            "operation_time": <operation time>
        }
}
```

### POST /internal/task

```json
{
    "id": <unique task id>,
    "result": <result>
}
```

## Особая благодарность

[C418 - Aria Math - The Dominator Synthwave Remix](https://www.youtube.com/watch?v=yiS0DPekSDQ)

[Undertale Yellow - END OF THE LINE_ - SayMaxWell Remix](https://www.youtube.com/watch?v=c54WQTqlFGU)

[John Williams - Setting the Trap - d.notive Synth Remix](https://www.youtube.com/watch?v=3zy-XqRXH1g)