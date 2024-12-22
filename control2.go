package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/api/v1/calculate", CalculateHandler)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	result, err := Calc(req.Expression)
	if err != nil {
		if err.Error() == "invalid expression" || err.Error() == "invalid character" || err.Error() == "mismatched parentheses" {
			http.Error(w, `{"error":"Expression is not valid"}`, http.StatusUnprocessableEntity)
		} else {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	res := Response{Result: fmt.Sprintf("%.2f", result)}
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("%s %s", time.Now().Format("2006/01/02 15:04:05"), jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Calc вычисляет математическое выражение, заданное в виде строки.
func Calc(expression string) (float64, error) {
	tokens := tokenize(expression)
	formattedSlice, err := ConvertToNumberNumberOperator(tokens)
	if err != nil {
		return 0, err
	}
	return CalculateNumberNumberOperatorToResult(formattedSlice)
}

// tokenize разбивает выражение на токены(части).
func tokenize(expr string) []string {
	var tokens []string
	var currentToken string

	for _, char := range expr {
		switch char {
		case ' ':
			continue
		case '+', '-', '*', '/', '(', ')':
			if len(currentToken) > 0 {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
			tokens = append(tokens, string(char))
		default:
			currentToken = string(char)
		}
	}
	//для добавления последнего
	if len(currentToken) > 0 {
		tokens = append(tokens, currentToken)
	}

	return tokens
}

// ConvertToNumberNumberOperator  преобразует формат выражения в Число Число Оператор и убирает скобки.
func ConvertToNumberNumberOperator(tokens []string) ([]string, error) {
	var output []string
	var operators []string

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, errors.New("Mismatched parentheses")
			}
			operators = operators[:len(operators)-1]
		} else if isOperator(token) {
			for len(operators) > 0 && priority(operators[len(operators)-1]) >= priority(token) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		} else {
			return nil, errors.New("invalid character")
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" {
			return nil, errors.New("mismatched parentheses")
		}

		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}
	return output, nil
}

// CalculateNumberNumberOperatorToResult вычисляет выражение.
func CalculateNumberNumberOperatorToResult(formattedSlice []string) (float64, error) {
	var tempResult []float64

	for _, token := range formattedSlice {
		if isNumber(token) {
			num, _ := strconv.ParseFloat(token, 64)
			tempResult = append(tempResult, num)
		} else if isOperator(token) {
			if len(tempResult) < 2 {
				return 0, errors.New("invalid expression")
			}
			b := tempResult[len(tempResult)-1]
			a := tempResult[len(tempResult)-2]
			tempResult = tempResult[:len(tempResult)-2]

			switch token {
			case "+":
				tempResult = append(tempResult, a+b)
			case "-":
				tempResult = append(tempResult, a-b)
			case "*":
				tempResult = append(tempResult, a*b)
			case "/":
				if b == 0 {
					return 0, errors.New("division by zero")
				}
				tempResult = append(tempResult, a/b)
			default:
				return 0, errors.New("unknown operator: " + token)
			}
		} else {
			return 0, errors.New("invalid token: " + token)
		}
	}

	if len(tempResult) != 1 {
		return 0, errors.New("invalid expression")
	}

	return tempResult[0], nil
}

// isNumber проверяет, является ли токен числом.
func isNumber(token string) bool {
	if _, err := strconv.ParseFloat(token, 64); err == nil {
		return true
	}
	return false
}

// isOperator проверяет, является ли токен оператором.
func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

// priority возвращает приоритет оператора.
func priority(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}
