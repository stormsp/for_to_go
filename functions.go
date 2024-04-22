package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"time"
	"math"
)

// TODO: nmin, nmax,dost, true,false, cyclesec,execsec,
type RepsValue interface{}

type Rep struct {
	sys_num int
	SYS_NUM int
	Value   bool
}

var Reps map[string]Rep

var database = map[string] bool {
	"a": true,
	"b": false,
}

//работа переводчика____________________________________________________

func findReps(text string) map[string]Rep {
	re := regexp.MustCompile(`\{([^{}\n]+)\}`)
	matches := re.FindAllStringSubmatch(text, -1)

	reps := make(map[string]Rep)

	for _, match := range matches {
		if len(match) > 1 {
			rep := match[1]
			// Убираем пробелы и символы табуляции в начале и конце строки
			rep = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(rep, "")

			// Проверяем, был ли репер добавлен ранее
			if _, found := reps[rep]; !found {
				// Генерируем случайное значение 0 или 1
				randomValue := rand.Intn(2)
				if randomValue == 0 {
					reps[rep] = Rep{Value: false}
				} else {
					reps[rep] = Rep{Value: true}
				}
			}
		}
	}

	return reps
}

func ReplaceExpressions(text string, reps map[string]Rep) string {
	re := regexp.MustCompile(`\{([^{}\n]+)\}`)

	// Заменяем выражения в тексте
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		repName := match[1 : len(match)-1] // Извлекаем имя репера из скобок
		if _, found := reps[repName]; found {
			return fmt.Sprintf("Reps[\"%s\"].Value", repName)
		}
		return match // Если репер не найден, оставляем выражение без изменений
	})

	return result
}

func ReplaceAllStringRegexp(input, pattern, replace string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllString(input, replace)
}

func ReplaceAllStringRegexpFunc(input, pattern string, repl func(string) string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllStringFunc(input, repl)
}

//работа переводчика____________________________________________________


//математические функции
//NMIN
func NMIN(values ...float64) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("no values provided")
	}
	min := math.Inf(1) // Инициализируем min как положительную бесконечность
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min, nil
}
//NMAX
func NMAX(values ...float64) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("no values provided")
	}
	max := math.Inf(-1) // Инициализируем max как отрицательную бесконечность
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max, nil
}

//логические функции
// DOST проверяет достоверность переменной по её имени
func DOST(varName string) int {
	if valid, exists := database[varName]; exists && valid {
		return 1
	}
	return 0
}
// TRUE всегда возвращает true, независимо от входных аргументов
func TRUE(args ...interface{}) bool {
	return true
}

// FALSE принимает любое количество аргументов и всегда возвращает false
func FALSE(args ...interface{}) bool {
	return false
}


//битовые функции


// BITS получает значение группы битов по маске, начиная с заданного бита
func BITS(dw uint32, bit0 uint, mask uint32) uint32 {
	return (dw >> bit0) & mask
}

// BXCHG переставляет байты в двойном слове в соответствии с заданной последовательностью
func BXCHG(dw uint32, byteseq string) uint32 {
	var result uint32
	for i, char := range byteseq {
		if char >= '1' && char <= '4' {
			shift := (4 - uint(char-'0')) * 8
			result |= ((dw >> shift) & 0xFF) << (3 - i) * 8
		}
	}
	return result
}

// SETBITS устанавливает значения битов в числе на заданное значение
func SETBITS(dw uint32, cnt uint, shf uint, val uint32) uint32 {
	mask := uint32((1<<cnt - 1) << shf) // Создаём маску для установки битов
	return (dw &^ mask) | ((val << shf) & mask)
}



func GETTICKS(prevTickCnt int) int {
	currentTickCnt := int(time.Now().UnixNano() / int64(time.Millisecond))
	if prevTickCnt != 0 {
		return currentTickCnt - prevTickCnt
	}
	return currentTickCnt
}

func TICKSIZE() int {
	// Засекаем начальное время
	startTime := time.Now()
	// Засекаем конечное время после прошедшей одной секунды
	time.Sleep(1 * time.Second)
	endTime := time.Now()
	// Рассчитываем разницу в секундах и преобразуем её в int
	duration := int(endTime.Sub(startTime).Seconds())
	// Выводим разницу в секундах
	fmt.Printf("TICKSIZE: %d seconds\n", duration)
	// Возвращаем значение
	return duration
}


func SET(parameter any, value any) {
	parameter = value
}

func SET_WAIT(parameter any, value any, timeout any) bool {
	parameter = value
	return true
}



// isBool проверяет, является ли значение логическим (bool)
func isBool(val RepsValue) bool {
	_, ok := val.(bool)
	return ok
}


// isInt проверяет, является ли значение целочисленным (int)
func isInt(val RepsValue) bool {
	_, ok := val.(int)
	return ok
}

// getBoolValue возвращает булево значение из RepsValue
func convertToInteger(value RepsValue) int {
	switch v := value.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return v
	default:
		// По умолчанию возвращаем 0 или другое значение по вашему усмотрению
		return 0
	}
}

func FINDOUT(first int, value int, count int, arr []int) int {
	for i := first; i < first+count; i++ {
		if arr[i] == value {
			return i
		}
	}
	return -1 // Возвращаем -1, если элемент не найден
}

func reset(param int) error {
	if param == -1 {
		// Выполнить операцию shutdown для Windows
		cmd := exec.Command("shutdown", "/r", "/t", "0")
		return cmd.Run()
	}

	// В противном случае, вам нужно определить логику для мягкой перезагрузки в Linux,
	// например, использование команды kill для отправки сигнала перезагрузки
	// или других подходящих методов.

	return fmt.Errorf("Unsupported operation")
}
