package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"time"
	"math"
	"runtime"
	"os"
)

// TODO: nmin, nmax,dost, true,false, cyclesec,execsec,
type RepsValue interface{}

type Rep struct {
	sys_num int
	SYS_NUM int
	Value   bool
}

var Reps map[string]Rep

var database = map[any] bool {
	"a": true,
	"b": false,
}

const (
	// Предположим, что период цикла в тиках таймера задан как константа
	ticksPerCycle uint = 100 // количество тиков в одном цикле

	// Количество тиков в секунде, предположим, что 1 секунда = 100 тиков
	ticksPerSecond uint = 100
)

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
func DOST(varName any) int {
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

//функции времени выполнения
// CYCLESEC возвращает период запуска алгоритма в секундах
func CYCLESEC() float64 {
	return float64(ticksPerCycle) / float64(ticksPerSecond)
}

// Функция, которую мы хотим замерить
func someTask() {
	// Имитация некоторой длительной операции
	time.Sleep(2 * time.Second)
}

// EXECSEC измеряет и возвращает время выполнения функции someTask в секундах
func EXECSEC() float64 {
	startTime := time.Now() // Засекаем время начала выполнения
	someTask()             // Выполнение функции, время которой необходимо измерить
	duration := time.Since(startTime) // Вычисляем длительность выполнения
	return duration.Seconds()         // Возвращаем длительность в секундах
}


//функции над таймерами
func TIMERMSEC(t time.Time) int {
	return t.Nanosecond() / 1e6
}
// TIMERSEC возвращает секунды от начала времени, указанного в параметре
func TIMERSEC(t time.Time) int {
	return t.Second()
}
// TIMERMIN возвращает минуты от начала времени, указанного в параметре
func TIMERMIN(t time.Time) int {
	return t.Minute()
}
// TIMERHOUR возвращает часы от начала времени, указанного в параметре
func TIMERHOUR(t time.Time) int {
	return t.Hour()
}
// MAKETIMER рассчитывает время счётчика из пользовательских данных
func MAKETIMER(hour, min, sec, msec int) time.Time {
	return time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, min, sec, msec*1e6, time.Local)
}


//функции счетчиков тиков
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

//функции перезагрузки не понял как реализовать на винде
// stopSoftdog имитирует остановку "программного сторожевого таймера".
// Теперь функция также проверяет, работает ли она на Linux, и только тогда выполняет свои действия.
func STOP_SOFTDOG() {
	// Получаем информацию об операционной системе
	osType := runtime.GOOS

	if osType != "linux" {
		fmt.Println("Функция STOP_SOFTDOG поддерживается только на Linux.")
		return
	}

	fmt.Println("STOP_SOFTDOG: Создание файла coredump.txt и завершение работы программы на Linux.")
	file, err := os.Create("coredump.txt")
	if err != nil {
		fmt.Println("Ошибка при создании файла coredump.txt:", err)
		return
	}
	defer file.Close()

	file.WriteString("Coredump caused by software watchdog timer.\n")

	// Эмулируем завершение работы программы
	os.Exit(1)
}

func RESET(param int) {
	switch runtime.GOOS {
	case "windows":
		if param == -1 {
			fmt.Println("Выполнение команды shutdown для Windows.")
			cmd := exec.Command("shutdown", "/s", "/t", "0") // Немедленное выключение
			if err := cmd.Run(); err != nil {
				fmt.Println("Ошибка при выполнении команды shutdown:", err)
			}
		} else {
			fmt.Println("Мягкая перезагрузка не поддерживается на Windows с параметром, отличным от -1.")
		}
	case "linux":
		fmt.Println("Создание файла coredump и мягкая перезагрузка на Linux.")
		_, err := os.Create("coredump.txt")
		if err != nil {
			fmt.Println("Не удалось создать файл coredump.txt:", err)
			return
		}
		fmt.Println("Файл coredump.txt создан успешно.")
		// Имитация деления на ноль для вызова паники
		fmt.Println("Деление на ноль для искусственной перезагрузки.")
		_ = 1 / (param - param) // Паника: деление на ноль
	default:
		fmt.Println("Операционная система не поддерживается.")
	}
}



//функции алгоритмов управления
func SET(parameter any, value any) {
	parameter = value
}

func SET_WAIT(parameter any, value any, timeout any) bool {
	parameter = value
	return true
}


//функции работы с массивами
func FINDOUT(first int, value int, count int, arr []int) int {
	for i := first; i < first+count; i++ {
		if arr[i] == value {
			return i
		}
	}
	return -1 // Возвращаем -1, если элемент не найден
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

func INITOUTS(firstIndex, value, count int) []int {
	// Создаем массив для хранения выходных переменных
	output := make([]int, count)

	// Инициализируем выходные переменные значениями value
	for i := 0; i < count; i++ {
		output[i] = value
	}

	return output
}

// Оператор BEEP - выдает звуковой сигнал (однократный короткий "бип").
func BEEP() {
	fmt.Println("Beep!")
}

// Оператор SIREN_ON - включает звуковой сигнал "сирена".
func SIREN_ON() {
	fmt.Println("Siren ON")
	// Ваш код для включения сирены
}

// Оператор SIREN_OFF - выключает звуковой сигнал "сирена".
func SIREN_OFF() {
	fmt.Println("Siren OFF")
	// Ваш код для выключения сирены
}

// Оператор EXECUTE - выполняет командный файл по указанному пути и имени.
func EXECUTE(path, filename string) error {
	cmd := exec.Command(path, filename)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
