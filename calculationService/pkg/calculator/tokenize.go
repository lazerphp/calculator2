package calc

import (
	"bytes"
	"errors"
	"slices"
	"strconv"
)

func PrepareExp(input string) ([]any, error) {

	var res []any           // результат
	var buf []rune          // массив для сборки числа
	var prev rune           // предыдущий не пробельный символ
	var balanceBrackets int // баланс скобок
	var cntDot int          // счетчик делителей в числе (надо <= 1)
	var insideNum bool      // флаг, что число в буфере собирается

	if len(input) == 0 {
		return nil, errors.New("пустой ввод")
	}

	for _, x := range input {
		if x == ' ' {
			insideNum = false
			if prev == '~' {
				return nil, errors.New("унарный минус стоит отдельно")
			}
			continue

		} else if x >= '0' && x <= '9' {
			if len(buf) == 0 && !slices.Contains([]rune{0, '(', '~', '*', '/', '+', '-'}, prev) {
				return nil, errors.New("цифра не на своем месте")
			}
			if len(buf) > 2 && buf[0] == '0' && buf[1] != '.' {
				return nil, errors.New("в числе ведущие нули")
			}
			if len(buf) > 0 && !insideNum {
				return nil, errors.New("число разбито")
			}
			insideNum = true
			buf = append(buf, x)

		} else if x == '.' {
			if !insideNum {
				return nil, errors.New("точка не в числе")
			}
			if cntDot > 0 {
				return nil, errors.New("больше 1 точки в числе")
			}
			cntDot += 1
			buf = append(buf, x)

		} else if slices.Contains([]rune{0, '+', '-', '*', '/'}, x) {
			switch x {
			case '-':
				if !(prev >= '0' && prev <= '9' || slices.Contains([]rune{0, '(', ')', '*', '/'}, prev)) {
					return nil, errors.New("неуместный знак -")
				}
				// унарный минус
				if prev == '*' || prev == '/' || prev == 0 {
					x = '~'
					continue
				}
			default:
				if !(prev >= '0' && prev <= '9' || prev == ')') {
					return nil, errors.New("неуместный оператор")
				}
			}

			if len(buf) > 0 {
				bufNum, _ := strconv.ParseFloat(string(buf), 64)
				res = append(res, bufNum)
				insideNum = false
				buf = nil
				cntDot = 0
			}
			res = append(res, x)

		} else if x == '(' || x == ')' {
			switch x {
			case '(':
				balanceBrackets++
				if !bytes.ContainsRune([]byte{0, '(', '~', '*', '/', '+', '-'}, prev) {
					return nil, errors.New("неуместная открывающая скобка")
				}
			case ')':
				balanceBrackets--
				if prev == '(' {
					return nil, errors.New("пустые скобки")
				}
				if !(prev == ')' || prev >= '0' && prev <= '9') {
					return nil, errors.New("неуместная закрывающая скобка")

				}
			}

			if balanceBrackets < 0 {
				return nil, errors.New("нарушена логика скобок")
			}

			if len(buf) != 0 {
				bufNum, _ := strconv.ParseFloat(string(buf), 64)
				res = append(res, bufNum)
				insideNum = false
				buf = nil
				cntDot = 0
			}
			res = append(res, x)

		} else {
			return nil, errors.New("невалидные символы")
		}

		prev = x
	}
	if len(buf) != 0 {
		bufNum, _ := strconv.ParseFloat(string(buf), 64)
		res = append(res, bufNum)
	}

	if !(prev >= '0' && prev <= '9' || prev == ')') {
		return nil, errors.New("незаконченное выражение")
	}
	if balanceBrackets != 0 {
		return nil, errors.New("лишние скобки")
	}

	return res, nil
}
