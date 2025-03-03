from http import HTTPStatus
import stat
from typing import List, Tuple
import requests

API = "http://localhost:8080/api/v1/"


class Expr:
    def __init__(self, ident, status, result):
        self.id = ident
        self.status = status
        self.result = result

    def __repr__(self):
        return (
            f"Expression {self.id}:\n - Status {self.status}\n - Result: {self.result}"
        )


def sendExpr(expression: str) -> Tuple[int, int]:
    x = requests.post(API + "calculate/", json={"expression": expression})
    if x.status_code == HTTPStatus.CREATED:
        return (x.json()["id"], x.status_code)
    else:
        print(x.content)
        return (-1, x.status_code)


def getExprs() -> Tuple[List[Expr], int]:
    x = requests.get(API + "expressions/")

    if x.status_code == HTTPStatus.OK:
        res = x.json()["expressions"]
        return (
            [Expr(r["id"], r["status"], r["result"]) for r in res] if res else [],
            x.status_code,
        )
    else:
        return ([], x.status_code)


def getExpr(id: int) -> Tuple[Expr, int]:
    x = requests.get(API + f"expressions/{id}")

    if x.status_code == HTTPStatus.OK:
        res = x.json()["expression"]
        return (Expr(res["id"], res["status"], res["result"]), x.status_code)
    else:
        return (None, x.status_code)


def do_op():
    print("Choose operation:")
    print(" [0] - Send Expression (Отправка)")
    print(" [1] - Get Expressions (Все задачи)")
    print(" [2] - Get Expression (Конкретную задачу)\n")

    i = input("Select operation: ")

    try:
        i = int(i)
    except:
        print("ERR: Invalid input")
        return

    match i:
        case 0:
            expr = input("Enter expression: ")
            print("Sending...")
            result = sendExpr(expr)
            match result[1]:
                case HTTPStatus.CREATED:
                    print(f"Created, id: {result[0]}")
                    return
                case HTTPStatus.UNPROCESSABLE_ENTITY:
                    print("Bad expression: ")
                    return
                case HTTPStatus.INTERNAL_SERVER_ERROR:
                    print("Internal Server Error")
                    return
                case _:
                    print(f"Unknown status: {result[1]}")
                    return
        case 1:
            print("Sending...")
            result = getExprs()
            match result[1]:
                case HTTPStatus.OK:
                    print(f"Recieved:")
                    for expr in result[0]:
                        print(expr)
                    return
                case HTTPStatus.INTERNAL_SERVER_ERROR:
                    print("Internal Server Error")
                    return
                case _:
                    print(f"Unknown status: {result[1]}")
                    return
        case 2:
            result = getExpr(input("Expression ID: "))
            print("Sending...")
            match result[1]:
                case HTTPStatus.OK:
                    print(f"Recieved:")
                    print(result[0])
                    return
                case HTTPStatus.NOT_FOUND:
                    print("Not found")
                case HTTPStatus.INTERNAL_SERVER_ERROR:
                    print("Internal Server Error")
                    return
                case _:
                    print(f"Unknown status: {result[1]}")
                    return


go = True
while go:
    do_op()
    go = input("Type x to exit, enter to stay: ").lower() != "x"
