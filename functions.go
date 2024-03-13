package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"time"
)

// TODO: nmin, nmax,dost, true,false, cyclesec,execsec,
type RepsValue interface{}

type Rep struct {
	sys_num int
	SYS_NUM int
	Value   bool
}

var Reps map[string]Rep

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

func replaceExpressions(text string, reps map[string]Rep) string {
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

func replaceAllStringRegexp(input, pattern, replace string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllString(input, replace)
}

func replaceAllStringRegexpFunc(input, pattern string, repl func(string) string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllStringFunc(input, repl)
}

//работа переводчика____________________________________________________

// BIT - получить значение бита из двойного слова
func BIT(dw uint32, bit0 uint) uint32 {
	return (dw >> bit0) & 1
}

// BITS - получить значение группы битов двойного слова по маске
func BITS(dw uint32, bit0 uint, mask uint32) uint32 {
	return (dw >> bit0) & mask
}

// BXCHG - переставить байты в двойном слове согласно заданной последовательности
func BXCHG(dw uint32, byteseq string) uint32 {
	// Здесь может быть реализация, учитывающая последовательность byteseq для перестановки байтов
	// Пример реализации может потребовать дополнительного кода для парсинга byteseq
	return dw // Заглушка для примера
}

// SETBITS - заменить cnt бит в dw, начиная с позиции shf, на значение val
func SETBITS(dw uint32, cnt uint, shf uint, val uint32) uint32 {
	mask := uint32((1<<cnt)-1) << shf // Убедимся, что маска имеет тип uint32
	return (dw &^ mask) | ((val&(1<<cnt) - 1) << shf)
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

func DOST(a any) any {
	if a == a {
		return 1
	}
	return 1
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
