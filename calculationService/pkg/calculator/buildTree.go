package calc

import (
	"fmt"
)

type Tree struct {
	nodeType string
	value    any
	left     *Tree
	right    *Tree
}

type bracketsCheck struct {
	counter int
}

// бинарные операторы в возрастающей приоритетности
// унарный минус рассматривается отдельно
var binaryOperators = [][]rune{
	{'-', '+'},
	{'*', '/'},
}

func findBinaryOperator(arr []any, operators ...rune) (int, bool) {
	start := 0
	end := len(arr)
	brackets := bracketsCheck{}

	for i := end - 1; i >= start; i-- {
		if arr[i] == ')' {
			brackets.counter++
		}
		if arr[i] == '(' {
			brackets.counter--
		}

		if brackets.counter == 0 {
			for _, op := range operators {
				if arr[i] == op {
					return i, true
				}
			}
		}
	}
	return 0, false
}

func BuildTree(arr []any) *Tree {

	start := 0
	end := len(arr)

	// ищем и убираем пары скобок вокруг
	for arr[start] == '(' {
		cnt, step := 0, 0
		for i := range len(arr) {
			if arr[i] == '(' {
				cnt++
			}
			if arr[i] == ')' {
				cnt--
			}
			if cnt == 0 {
				break
			}
			step++
		}
		if step == end-1 {
			start++
			end--
		} else {
			break
		}
	}

	// поиск операторов с низшего уровня с конца и обработка
	for _, ops := range binaryOperators {
		indexOperator, ok := findBinaryOperator(arr[start:end], ops...)
		if ok {
			indexOperator += start
			res := &Tree{}
			res.nodeType = "operator"
			res.value = arr[indexOperator]
			res.left = BuildTree(arr[start:indexOperator])
			res.right = BuildTree(arr[indexOperator+1 : end])
			return res
		}
	}

	// унарный минус
	if arr[start] == '~' {
		res := &Tree{}
		res.nodeType = "operator"
		res.value = '-'
		res.left = &Tree{"num", 0.0, nil, nil}
		res.right = BuildTree(arr[start+1 : end])
		return res
	}

	// если дошли досюда, значит это число
	res := &Tree{}
	res.nodeType = "num"
	num := arr[start]
	switch num := num.(type) {
	case int:
		res.value = float64(num)
	default:
		res.value = num
	}

	return res
}

// функция для простого вывода дерева
//
// принимает дерево *Tree и число отступов. При вызове cnt ставьте 0, мне лень делать обертку.
//
// используйте для отладки
func WalkTree(t *Tree, cnt int) {

	for range cnt {
		fmt.Print("|--")
	}

	var value any
	switch t.value.(type) {
	case rune:
		value = string(t.value.(rune))
	case float64:
		value = t.value.(float64)

	}
	fmt.Println("{", t.nodeType, value, "}")

	cnt++
	if t.left != nil {
		WalkTree(t.left, cnt)
	}

	if t.right != nil {
		WalkTree(t.right, cnt)
	}
}
