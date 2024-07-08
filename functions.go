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
	"bufio"
	"strings"
)

// TODO: nmin, nmax,dost, true,false, cyclesec,execsec,
type RepsValue interface{}

type Rep1 struct {
	sys_num int
	SYS_NUM int
	Value   bool
}

//var Reps map[string]Rep

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
					reps[rep] = Rep{Value: 0}
				} else {
					reps[rep] = Rep{Value: 1}
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
func NMIN(values ...float32) (float32, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("no values provided")
	}
	min := float32(math.Inf(1)) // Инициализируем min как положительную бесконечность
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min, nil
}
//NMAX
func NMAX(values ...float32) (float32, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("no values provided")
	}
	max := float32(math.Inf(-1)) // Инициализируем max как отрицательную бесконечность
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max, nil
}

//логические функции
// DOST проверяет достоверность переменной по её имени
func DOST(varName any) bool {
	if valid, exists := database[varName]; exists && valid {
		return true
	}
	return false
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
func CYCLESEC() float32 {
	return float32(ticksPerCycle) / float32(ticksPerSecond)
}

// Функция, которую мы хотим замерить
func someTask() {
	// Имитация некоторой длительной операции
	time.Sleep(2 * time.Second)
}

// EXECSEC измеряет и возвращает время выполнения функции someTask в секундах
func EXECSEC() float32 {
	startTime := time.Now() // Засекаем время начала выполнения
	someTask()             // Выполнение функции, время которой необходимо измерить
	duration := time.Since(startTime) // Вычисляем длительность выполнения
	return float32(duration.Seconds())      // Возвращаем длительность в секундах
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
var vars []float32
func pidreg(
	boi int,
	S0_Ki, S1_Kp, S2_Kd float32,
	g, fu, y, y_diap, umin, umax, Tf, u_vmax, g_vmax float32,
	reg_mode int,
	yf_size int,
	yf_type, adapt_type, gf_type, def_type, y_trust int) float32 {
	// Constants
	const tickRate = 0.01 // Tick rate in seconds

	// Initialize internal variables if not already initialized
	if len(vars) < boi+27 {
		newVars := make([]float32, boi+27)
		copy(newVars, vars)
		vars = newVars
	}

	// Update tick counter
	vars[boi+0] += tickRate

	// Calculate filtered setpoint
	if gf_type == 1 {
		// Limit rate of change for setpoint
		vars[boi+2] += float32(math.Min(float64(g_vmax)*float64(tickRate), math.Abs(float64(g)-float64(vars[boi+2]))) * math.Copysign(1, float64(g)-float64(vars[boi+2])))
	} else if gf_type == 2 {
		// Exponential smoothing
		vars[boi+2] += float32((float32(g) - vars[boi+2]) * float32(1 - math.Exp(float64(float32(-tickRate)/Tf))))
	} else {
		// No filter
		vars[boi+2] = g
	}

	// Calculate filtered feedback signal
	if yf_type == 1 {
		// Median filter (for simplicity, nearest to the average is used)
		vars[boi+3] = (vars[boi+3] + y) / 2
	} else if yf_type == 2 {
		// Simple moving average
		vars[boi+18] = (vars[boi+18]*float32(yf_size-1) + y) / float32(yf_size)
		vars[boi+3] = vars[boi+18]
	} else {
		// No filter
		vars[boi+3] = y
	}

	// Calculate error
	vars[boi+4] = vars[boi+2] - vars[boi+3]

	// Adaptation of coefficients
	if adapt_type == 1 && math.Abs(float64(vars[boi+4]/y_diap)) < 0.01 {
		S0_Ki *= 0.5
		S1_Kp *= 0.1
		S2_Kd *= 0.1
	}

	// Calculate PID terms
	vars[boi+9] = S1_Kp * vars[boi+4]                     // Proportional term
	vars[boi+10] = vars[boi+12] + S0_Ki * vars[boi+4]*tickRate // Integral term
	vars[boi+11] = S2_Kd * (vars[boi+4] - vars[boi+5]) / tickRate // Derivative term

	// Update integral component
	vars[boi+12] = vars[boi+10]

	// Calculate control signal
	u := vars[boi+9] + vars[boi+10] + vars[boi+11]

	// Apply output limits
	if u > umax {
		u = umax
	} else if u < umin {
		u = umin
	}

	// Apply rate of change limits
	if u_vmax > 0 {
		u += float32(math.Min(float64(u_vmax*tickRate), math.Abs(float64(u-vars[boi+14]))) * math.Copysign(float64(1), float64(u-vars[boi+14])))
	}

	// Save previous error for next derivative calculation
	vars[boi+5] = vars[boi+4]

	// Update internal state
	vars[boi+14] = u

	// Return control signal
	return u
}


//функции алгоритмов управления
func SET(parameter *float32, value float32) {
	*parameter = value
}

func SET_WAIT(parameter *float32, value float32, timeout any) bool {
	//fmt.Println(*parameter, float32(value))
	*parameter = value
	//fmt.Println(*parameter, value)
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


// Функция для добавления фигурных скобок к конструкциям if
func addBracesToIfStatements(code string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(code))

	// Регулярное выражение для поиска строк, начинающихся с if
	reIf := regexp.MustCompile(`^\s*if\s*\(?.*\)?\s*$`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		// Проверяем, начинается ли строка с if и нет ли уже фигурной скобки
		if reIf.MatchString(trimmedLine) && !strings.HasSuffix(trimmedLine, "{") {
			line += " {"
		}
		result.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading text:", err)
	}

	return result.String()
}

//Библиотеки __________________________________________________________________________________________________________
//valTrack.evl
func valTrack(val any, timeout float32, id float32) bool {
	if (convertToInteger(val) == convertToInteger(0)) {
		aout[int(id)]=0
		return(false)
	}

	// aout[id] время перехода в состояние, отличное от 0
	// для вычисления тайм-аута (в тиках со старта зонда)
	if (convertToInteger(aout[int(id)]) == convertToInteger(0)) {
		aout[int(id)]=GETTICKS(0)
	}

	if (convertToInteger(GETTICKS(aout[int(id)])*TICKSIZE()) >= convertToInteger(timeout)) {
		return(true)
	}

	return(false)
}

//
// valTrackGt и valTrackLt возращают 0 если отслеживаемый
// параметр не достоверен, или если не нарушена
// граница, или если со времени нарушения
// не прошло timeout секунд. В противном случае
// функции возвращают 1.
//
func valTrackGt(val float32, bound float32, timeout float32, id float32) bool {
	return(valTrack(convertToInteger(DOST(val)) != convertToInteger(0) && (convertToInteger(val) > convertToInteger(bound)),timeout,id))
}

func valTrackLt(val any, bound any, timeout float32, id float32) bool {
	return(valTrack(convertToInteger(DOST(val)) != convertToInteger(0) && (convertToInteger(val) < convertToInteger(bound)),timeout,id))
}

// при достоверности одного из трех каналов, недостоверный канал заменяется
// значением параметров с ЭКМ давление на вых низкое, высокое
//
func valTrackLt_DOST(val any, bound any, timeout float32, id float32, p_ekm any) any {
	if DOST(val) {
		return(valTrack((convertToInteger(val) < convertToInteger(bound)),timeout,id))
	} else {
		return(p_ekm)
	}
}


func valTrackGt_DOST(val any, bound any, timeout float32, id float32, p_ekm any) any {
	if DOST(val) {
		return(valTrack((convertToInteger(val) > convertToInteger(bound)),timeout,id))
	} else {
		return(p_ekm)
	}
}


//	yahont.evl
func yahont4(a int, cod any) {
	if DOST(cod) {
		if (convertToInteger(cod) == convertToInteger(0)) || (convertToInteger(cod) == convertToInteger(1)) || (convertToInteger(cod) == convertToInteger(2)) {//неопр || кз || обрыв {
			dout[a]=0 //--неопр
		} else {
			if (convertToInteger(cod) == convertToInteger(3)) {//норма {
				dout[a]=1 //--норма
			} else {
				if (convertToInteger(cod) == convertToInteger(4)) {//вним {
					dout[a]=3
				} else {
					if (convertToInteger(cod) == convertToInteger(5)) {//трев {
						dout[a]=2 //--пожар
					} else {
						dout[a]=0 //--неисправн
					}
				}
			}
		}
	} else {
		dout[a]=0
	}
}


// —осто¤ние концевика
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-вскитые
//

func yahont41(a int, cod any) any {
	if (convertToInteger(cod) == convertToInteger(0)) {   		// код 0-неопр {
		dout[a]=0
	} else {
		if (convertToInteger(cod) == convertToInteger(3)) {              // код 3-норма {
			dout[a]=1
		} else {
			if (convertToInteger(cod) == convertToInteger(6)) {  		// код 6-трев {
				dout[a]=2
			} else {
				dout[a]=0		// код cбой
			}
		}
	}
	return(false)
}


// яхонт4 - 3 расч статус источника питани¤ (неопр,норма,неиспр)
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-нет питани¤
//

func yahont42(a int, cod any) any {
	if (convertToInteger(cod) == convertToInteger(0)) {   		// код 0-неопр {
		dout[a]=0
	} else {
		if (convertToInteger(cod) == convertToInteger(3)) {              // код 3-норма {
			dout[a]=1
		} else {
			if (convertToInteger(cod) == convertToInteger(6)) {  		// код 6-трев {
				dout[a]=2
			} else {
				dout[a]=0		// остальное cбой
			}
		}
	}
	return(false)
}


// -------------------------- яхонт16и -------------------------------
// —осто¤ние ѕќ∆ј–Ќџ’ шлейфов
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-пожар, 3-внимание
//

func yahont16_ps(a int, cod any) any {
	if DOST(cod) {
		if (convertToInteger(cod) == convertToInteger(0)) || (convertToInteger(cod) == convertToInteger(1)) || (convertToInteger(cod) == convertToInteger(2)) {   		// 0-неопр// 1-кз 2-обрыв 6-обрыв {
			dout[a]=0                 // ** неопр
		} else {
			if (convertToInteger(cod) == convertToInteger(3)) {  // 3-норма {
				dout[a]=1               // ** норма
			} else {
				if (convertToInteger(cod) == convertToInteger(4)) {   // 4-вним {
					dout[a]=3
				} else {
					if (convertToInteger(cod) == convertToInteger(5)) {  	// 5-трев {
						dout[a]=2             // ** пожар/дым
					} else {
						dout[a]=0		// ** неисправн
					}
				}
			}
		}
	} else {
		dout[a]=0
	}
	return(false)
}


// —осто¤ние ќ’–јЌЌџ’ шлейфов
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-вскитые, 3- сн¤т с охраны
//

func yahont16_os(a int, cod any) any {
	if (convertToInteger(cod) == convertToInteger(135)) {   		// 0-неопр {
		dout[a]=0
	} else {
		if (convertToInteger(cod) == convertToInteger(132)) {              // 3-норма {
			dout[a]=1
		} else {
			if (convertToInteger(cod) == convertToInteger(134)) {  		// 6-трев {
				dout[a]=2
			} else {
				if (convertToInteger(cod) == convertToInteger(129)) || (convertToInteger(cod) == convertToInteger(130)) || (convertToInteger(cod) == convertToInteger(131)) {
					dout[a]=3		// 6-сн¤т
				} else {
					dout[a]=0		// cбой
				}
			}
		}
	}
	return(false)
}


// --- —осто¤ние источника питани¤ ---
// 0 - норма, 1- неисправность


func yahont16_pit(a int, cod int) {
	//2 переменные
	dout[a+0]=convertToInteger(convertToInteger(cod * 1) != convertToInteger(0))
	dout[a+1]=convertToInteger(convertToInteger(cod * 256) != convertToInteger(0))
}

//bupg24_3.evl
// БУПГ24-3, БУПГ24-6
// Галеев 08.2019

//----------- <Адрес 11>. Состояние дискретных датчиков -------------------
//Единицы в битах регистра (кроме бита 10) означают, что соответствующие датчики  находятся в состоянии "Авария"
// нули - что соответствующие датчики находятся в состоянии "Норма".
//
// Соответствие битов регистра входным сигналам:
//   0 - неисправность датч Твх - БУПГ24-6
//   1 - не используются//
//   2 - сигнал перегрева//
//   3 - сигнал <Давление газа высокое>//
//   4 - сигнал <Давление газа низкое>//
//   5 - сигнал <Давление продукта высокое>//
//   6 - сигнал <Давление продукта низкое>//
//   7 - сигнал <Разрежение низкое>//
//   8 - сигнал <Уровень теплоносителя низкий>//
//   9 - сигнал <Прорыв газа>//
//   10 - сигнал <>// 1 - наличие пламени, 0 - его отсутствие.
//   11 - сигнал <Расход продукта низкий>//
//   12 - сигнал <Давление запальника высокое>//
//   13 - сигнал <Загазованность>//
//   14 - сигнал <Неисправность аналогового датчика  температуры теплоносителя>//
//   15 - сигнал <Неисправность аналогового датчика температуры газа на выходе>.
//-----------------------------
//--------<Адрес 12>. Аварийные состояния датчиков
//Соответствие битов регистра входным сигналам: аналогично регистрам 1, 11.
//При аварийном отключении подогревателя по какому-либо сигналу бит,
//соответствующий этому сигналу, устанавливается в 1.
//Содержимое регистра сохраняется до следующего аварийного отключения подогревателя или отключения питания БУПГ.

func bupg243(bp int, cod float32) int {
	codInt := int(cod)
	dout[bp+0] = ne(codInt&4, 0)     // Перегрев
	dout[bp+1] = ne(codInt&8, 0)     // Ргаза высокое
	dout[bp+2] = ne(codInt&16, 0)    // Ргаза низкое
	dout[bp+3] = ne(codInt&32, 0)    // Ртн высокое
	dout[bp+4] = ne(codInt&64, 0)    // Ртн низкое
	dout[bp+5] = ne(codInt&128, 0)   // Разряжение низкое
	codInt = codInt / 256
	dout[bp+6] = ne(codInt&1, 0)     // Уровень тн низкий
	dout[bp+7] = ne(codInt&2, 0)     // Прорыв газа
	dout[bp+8] = ne(codInt&4, 0)     // Пламя 0-нет 1-есть
	dout[bp+9] = ne(codInt&8, 0)     // Расход низкий
	dout[bp+10] = ne(codInt&16, 0)   // Давление запальника высокое
	dout[bp+11] = ne(codInt&32, 0)   // Загазованность
	dout[bp+12] = ne(codInt&64, 0)   // Пламя отсутств(3М), Неиспр датч температуры тн
	dout[bp+13] = ne(codInt&128, 0)  // Неиспр датч температуры(3М) вых газа
	return 0
}

func bupg243_klap(bp int, cod float32) int {
	codInt := int(cod)
	dout[bp+0] = ne(codInt&2, 0)     // клапан запальника
	dout[bp+1] = ne(codInt&4, 0)     // клапан отсекателя
	dout[bp+2] = ne(codInt&8, 0)     // клапан б.горения
	dout[bp+3] = ne(codInt&16, 0)    // сигнал аварии
	dout[bp+4] = ne(codInt&32, 0)    // звук сигнала аварии
	dout[bp+5] = ne(codInt&64, 0)    // клапан безопасности
	return 0
}

func bupg(bp int, mode float32, cod11, cod12 float32) int {
	modeInt := int(mode)
	if eq(modeInt, 6) == 1 { // Авария
		bupg243(bp, cod12)
	} else {
		bupg243(bp, cod11)
	}
	return 0
}

//sgoes.evl
func eq(a, b int) int {
	if a == b {
		return 1
	}
	return 0
}

// sgoes_avar возвращает 1 если авария, иначе 0
func sgoes_avar(cod int) int {
	x := ne(cod&1, 1) // авария
	return x
}

// sgoes_porog1 возвращает 1 если превышен порог 1, иначе 0
func sgoes_porog1(cod int) int {
	x := eq(cod&2, 2) // порог 1 превышен
	return x
}

// sgoes_porog2 возвращает 1 если превышен порог 2, иначе 0
func sgoes_porog2(cod int) int {
	x := eq(cod&4, 4) // порог 2 превышен
	return x
}

// sgoes возвращает 1 если авария, иначе 0 (для совместимости со старым кодом САУ)
func sgoes(cod int) int {
	x := ne(cod&1, 1) // авария
	return x
}


//BOM2.evl
func ne(a, b int) int {
	if a != b {
		return 1
	}
	return 0
}

func bomOven(bp int, cod1, cod2 float32) {
	cod1Int := int(cod1)
	cod2Int := int(cod2)

	dout[bp+0] = ne(cod1Int & 2, 0)        // Клапан заправки
	dout[bp+1] = ne(cod1Int & 4, 0)        // Клапан пульсатора
	dout[bp+2] = ne(cod1Int & 16, 0)       // Клапан сброса
	dout[bp+3] = ne(cod1Int & 32, 0)       // Проникновение в одоризатор
	//cod := cod1Int / 256
	dout[bp+4] = ne(cod1Int & 2, 0) + 2*ne(cod1Int & 4, 0)  // Низкая/Высокая температура
	dout[bp+5] = ne(cod1Int & 8, 0)                         // Высокое давление в коллекторе
	dout[bp+6] = ne(cod1Int & 16, 0)                        // Высокий перепад давления
	dout[bp+7] = ne(cod1Int & 32, 0) + 2*ne(cod1Int & 64, 0) // Низкий/Высокий уровень в расходной емкости
	dout[bp+8] = ne(cod1Int & 128, 0)                       // Ошибка выдачи дозы

	dout[bp+9] = ne(cod2Int & 1, 0)         // Авария
	dout[bp+10] = ne(cod2Int & 2, 0)        // Пожар
	dout[bp+11] = ne(cod2Int & 4, 0)        // Обрыв датчика потока
	dout[bp+12] = ne(cod2Int & 8, 0)        // Неисправность датчика давления в коллекторе
	dout[bp+13] = ne(cod2Int & 16, 0)       // Неисправность датчика давления в емкости
	dout[bp+14] = ne(cod2Int & 32, 0)       // Неисправность сигнализатора уровня
	dout[bp+15] = ne(cod2Int & 64, 0)       // Неисправность датчика температуры
	//cod = cod2Int / 256
	dout[bp+16] = ne(cod2Int & 1, 0)        // РИП норма
	dout[bp+17] = ne(cod2Int & 2, 0)        // РИП батарея норма
	dout[bp+18] = ne(cod2Int & 8, 0)        // Пульсатор
}

//vbp.evl
// Galeev
// Расчет обьема газа при работе на байпасе
// при старте зонда обнулить переменные
// 5 переменных
//aout[ba+0] - время открытия
//aout[ba+1] - время закрытия
//aout[ba+2] - время работы на байпасе, час
//aout[ba+3] - объем газа на байпасе
//aout[ba+4] - время работы на байпасе, сек
// kr_bp - положение байпасного крана
// Vpsut - объем газа за прошлые сутки
// ba - базовый адрес

func Vbp(kr_bp float32, Vpsut float32, ba int) {
	if convertToInteger(kr_bp) == convertToInteger(1) {
		if DOST(aout[ba+1]) {
			aout[ba]=int(time.Now().Unix())   // время открытия
			aout[ba+1]=0
		}
		aout[ba+2]=(int(time.Now().Unix())-aout[ba])/3600 // в часах
		aout[ba+4]=(int(time.Now().Unix())-aout[ba])      // в сек
		aout[ba+3]=int(float32(aout[ba+2])*Vpsut/24)
	}

	if (convertToInteger(kr_bp) == convertToInteger(2)) && (convertToInteger(DOST(aout[ba+1])) == convertToInteger(0)) {
		aout[ba+1]=int(time.Now().Unix())
		aout[ba+2]=(aout[ba+1]-aout[ba])/3600
		aout[ba+4]=(aout[ba+1]-aout[ba])     // в сек
	}
}

func DOSTacc(sum float32, v1 float32) float32 {
	if DOST(v1) {
		sum=sum+v1
	}
	return(sum)
}

func pDOSTnearest(med float32, v1 float32, v2 float32) float32 {
	if DOST(v1) {
		if (convertToInteger(math.Abs(float64(med-v1))) > convertToInteger(math.Abs(float64(med-v2)))) {
			return(v2)
		}
		return(v1)
	}
	return(v2)
}

//2is3.evl
// возвращает индекс ближайшего из двух к среднему

// val,i - значение и индекс измерения
// DOSTval,di - значение и индекс достоверного измерения
// если val достоверно, возвращает индекс ближайшего из двух к среднему
// иначе водвращает di
//
func iDOSTnearest(med float32, v1 float32, i1 float32, v2 float32, i2 float32, ii float32) float32 {
	if DOST(v1) && DOST(v2) {
		if (convertToInteger(math.Abs(float64(med-v1))) > convertToInteger(math.Abs(float64(med-v2))+0.002)) { //+0.002 - чтобы убрать дребезг {
			return(i2)
		} else {
			return(i1)
		}
	} else {
		return(ii)
	}
}

// расчет индекса датчика давления для регулятора на двух эр-04
// 0 - эр04-12, 1 - эр04-21, 2 - эр04-22, подача в том же порядке
//
func regp3i(p1 float32, p2 float32, p3 float32, self float32) float32 {
	i:=float32(0)
	c:=float32(0)
	p:=float32(0)

	if DOST(p1) {
		c=c+1
		p=p+p1
	}


	if DOST(p2) {
		i=1
		c=c+1
		p=p+p2
	}

	if DOST(p3) {
		i=2
		c=c+1
		p=p+p3
	}


	if c != 0 {
		p=p/c
		i=iDOSTnearest(p,p1,0,p3,2,i)
		i=iDOSTnearest(p,p2,1,p1,0,i)
		i=iDOSTnearest(p,p3,2,p2,1,i)
	} else {
		i=self
	}

	return(i)
}


// выбор значения Рвых для регулятора по трем ан.датчикам
// pself - текущее значение параметра
//
func regp3p(p1 float32, p2 float32, p3 float32, self float32) any {
	psum:=DOSTacc(0,p1)
	psum=DOSTacc(psum,p2)
	psum=DOSTacc(psum,p3)

	c := 0
	if DOST(p1) {
		c = 1
	}
	p:=p1

	if DOST(p2) {
		c=c+1
		p=p2
	}

	if DOST(p3) {
		c=c+1
		p=p3
	}

	if c != 0 {
		p=pDOSTnearest(psum/float32(c),p1, p)
		p=pDOSTnearest(psum/float32(c),p2, p)
		p=pDOSTnearest(psum/float32(c),p3, p)
	} else {
		p=0
	}
	return(p)
}




// ПС 2 из 3 меньше, если достоверен только 1 датчик, не вырабатывать
// mux - множитель типа 90%
// тратит 3 переменные
//

func ps2is3Lt(p1 float32, p2 float32, p3 float32, mux float32, pzad float32, T float32, vi float32) bool {
	a1:= 0
	if (valTrackLt(p1,0.01*mux*pzad,T,vi)){
		a1 = 1
	}
	a2:= 0
	if (valTrackLt(p2,0.01*mux*pzad,T,vi+1)) {
		a2 = 1
	}
	a3:=0 
	if (valTrackLt(p3,0.01*mux*pzad,T,vi+2)) {
		a3 = 1
	}
	if (convertToInteger(DOST(p1)||DOST(p3)||DOST(p3)) >= convertToInteger(2)) {
		return ((convertToInteger(a1+a2+a3) >= convertToInteger(2)))
	}
	return(false)
}

//// ПС 2 из 3 больше, если достоверен только 1 датчик, не вырабатывать
// mux - множитель типа 110%
// тратит 3 переменные
//
func ps2is3Gt(p1 float32, p2 float32, p3 float32, mux float32, pzad float32, T float32, vi float32) bool {
	a1:= 0
	if (valTrackGt(p1,0.01*mux*pzad,T,vi)){
		a1 = 1 
	}
	a2:= 0
	if (valTrackGt(p2,0.01*mux*pzad,T,vi+1)){
		a2 = 1
	}
	a3:=0 
	if (valTrackGt(p3,0.01*mux*pzad,T,vi+2)){
		a3 = 1
	}
	if (convertToInteger(DOST(p1)||DOST(p3)||DOST(p3)) >= convertToInteger(2)) {
		return ((convertToInteger(a1+a2+a3) >= convertToInteger(2)))
		//return(a3)
	}
	return(false)
}


//uug.evl
// Расчет объемов QY (за прошлые сутки), QD (с начала суток), за прошлый месяц для sevc и SF

// ********** Функция вычисления qd, qy
// t      - время устройства на этом шаге
// vs_sys - vsum, параметр непрерывный расход (исходный для всех расчетов)
// qf_sys - fix, параметр, где фиксируется непрер накопленный расход
//          при смене контр часа (уст извне, чтобы хранить)
// qy_sys - sys параметра qy (уст извне, чтобы хранить)
// qd_ind - индекс расчетной переменной qd
// qmax   - максимальный возможный расход за сутки, больше - недост
// chour  - контрактный час
// aout[vi+0] - расход с начала суток
// aout[vi+1] - для слежения за изменением времени
//
func hour(t time.Time) int {
	return t.Hour()
}

// updQyQd updates qy and qd based on the given parameters.
func upd_qyqd(t1 *float32, vsSys, qfSys, qySys *float32, vi int, qmax float32, chour int) {
	// Check if contract hour has arrived
	t := time.Now()
	if hour(t) != hour(time.Unix(int64(aout[vi+1]), 0)) && hour(t) == chour {
		qy := *vsSys - *qfSys
		if qy < 0 || qy > qmax {
			qy = 0 // Set qy to 0 if it's out of valid range
		}
		*qySys = qy   // Update qySys with the current accumulated value
		*qfSys = *vsSys // Store the current accumulated value
		time.Sleep(2 * time.Second) // Sleep for 2 seconds to simulate usage of the new value
	}
	aout[vi] = int(*vsSys - *qfSys) // Calculate consumption since the start of the day
	aout[vi+1] = int(float32(t.Unix())) // Update the last time
}


// ********** Функция вычисления qm - расхода за месяц
// t      - время устройства на этом шаге
// vs_sys - vsum, параметр непрерывный расход (исходный для всех расчетов)
// qf_sys - fix, параметр, где фиксируется непрер накопленный расход
//          на начало месяца (уст извне, чтобы хранить)
// qm_sys - sys параметра qy (уст извне, чтобы хранить)
// vi     - индекс переменной для расчетов
// qmax   - максимальный возможный расход за месяц, больше - недост
// chour  - контрактный час
// dout[vi+0] - для управления расчетом
// aout[vi+1] - для слежения за изменением времени
//
// month extracts the month part from the given time.
func month(t time.Time) int {
	return int(t.Month())
}

// updQmes updates qm based on the given parameters.
func updQmes(t time.Time, vsSys, qfSys, qmSys *float32, vi int, qmax float32, chour int, dout, aout []float32) {
	// Check if a new month has started
	if month(t)-1 != month(time.Unix(int64(aout[vi+1]), 0))-1 {
		dout[vi] = 1 // Enable delayed calculation of qm
	}

	// Check if the calculation is enabled and the contract hour has arrived
	if dout[vi] == 1 && hour(t) != hour(time.Unix(int64(aout[vi+1]), 0)) && hour(t) == chour {
		qm := *vsSys - *qfSys
		if qm < 0 || qm > qmax {
			qm = 0 // Set qm to 0 if it's out of valid range
		}
		*qmSys = qm  // Update qmSys with the current accumulated value
		*qfSys = *vsSys // Store the current accumulated value
		time.Sleep(2 * time.Second) // Sleep for 2 seconds to simulate usage of the new value
		dout[vi] = 0 // Reset the calculation flag
	}

	aout[vi+1] = float32(t.Unix()) // Update the last time
}

	// переменные
	// 1,2 - upd_qyqd sd
	// 3,4 - upd_qmes sd
	// 5,6 - upd_qmes sf2et1
	// 7,8 - upd_qmes sf2et2

	// oninit(t)
	//aout[2]=Reps["SVC ВРЕМЯ БЕЛ"].Value        // чтобы замечать смену суток
	//aout[4]=Reps["SVC ВРЕМЯ БЕЛ"].Value        // чтобы замечать смену суток
	//aout[6]=Reps["БЕЛБ SF1-TIME"].Value         // чтобы замечать смену суток
	//aout[8]=Reps["БЕЛБ SF2-TIME"].Value         // чтобы замечать смену суток
	//aout[10]=Reps["БЕЛБ SF3-TIME"].Value         // чтобы замечать смену суток


func convertToBool(val float32) bool {
	return val != 0
}

//#include "eval.lib\set.evl"

// 01.06.15
// для вставки #include "eval.lib\set.evl"

//
// управление при условии достоверности
//
func setex(sys *float32, value float32) bool {
	if (convertToInteger(DOST(sys)) == convertToInteger(0)) {
		return(false)
	}
	SET(sys, value)
	return(true)
}

//
// setwex - аналог встроенной SET_WAIT
// однако, в случае не успеха
// производится дополнительные 1 попытки
// достигнуть заданного соcтояния
//
func setwex(sys *float32, value float32, timeout any) bool {
	//fmt.Println(*sys)
	if convertToInteger(SET_WAIT(sys,value,timeout)) != convertToInteger(0) {
		//fmt.Println("voshel")
		//fmt.Println(*sys)
		//time.Sleep((18) * time.Second)
		return(SET_WAIT(sys,value,timeout))
	}
	return(false)
}

//
// impuls
//
func impuls(sys *float32, t any) any {
	x:=SET_WAIT(sys,1,t)
	//time.Sleep((2*18) * time.Second)
	x=SET_WAIT(sys,0,t)
	return(x)
}



// установка значения с заданной чувствительностью
// возврат 1-установлено
//         0-без реакции
func setSens(sys *float32, value float32, sens any) bool {
	x := false
	if (convertToInteger(math.Abs(float64(*sys - value))) > convertToInteger(sens)) {
		x = setex(sys,value)
	}
	return x
}



func setwex_DOST(sys *float32, value float32, timeout any) any {
	if !DOST(sys) {
		return(false)
	}
	return(SET_WAIT(sys,value,timeout))
}









//#include "eval.lib\front.evl"
// front 0-> ne 0
// src - дискр сигнал
// id - номер переменной слежения
//
func front(src *float32, id int) bool {
	x := false
	if DOST(src) && convertToInteger(src) != convertToInteger(dout[id]) && convertToInteger(src) != convertToInteger(0) {
		x=true
	}
	dout[id]=int(*src)
	return(x)
}


// Тест КРБП
// Проверено в Телепаново
// Галеев 19.03.15

// 25.06.2015 :Галеев. Проверено в Таптыково
// 1.добавлена обработка события когда при начале теста положение крбп
//   сильно отличается от задания. это сразу дает неисправность
// 2.исключена ошибочная засылка задания выше 100%

//#INCLUDE "eval.lib\klap_test.evl"

//u=klap_test(u_1,Reps["РУЧПОЛБП ЯНАУ"].Value,Reps["ПОЛОЖЗАДВ ЯНЛ"].Value,33,Reps["КР БП ЯНАУ"].Value,7,Reps["РЕЖИМ ГРС"].Value)

// час текущего времени сау
//
func curhour () int {
	curtime:=time.Now()
	return(curtime.Hour())
}

//
// тест клапана
// man - ручное задание
// pol - положение клапана
// u - сигнал управления из pid
// dout[vi+0] - пс неисправности клапана
// aout[vi+1] - отсчет времени теста
// aout[vi+2] - предыдущий час
// dout[vi+3] - внеочередная проверка
//
func klap_test(u int, man float32, pol float32, vi float32, bp_kr any, t any, rejim_grs any) int {
	if (convertToInteger(bp_kr) == convertToInteger(2)) {

		if (convertToInteger(aout[int(vi+1.0)]) == convertToInteger(0)) {
			u=int(man)		// без проверки было бы так и все
		}

		h:=curhour()
		if convertToInteger(rejim_grs) != convertToInteger(0) {
			a := false
			if dout[int(vi +3)] > 0 {
				a = true
			}
			if convertToInteger(h) != convertToInteger(aout[int(vi+2)]) && (convertToInteger(h) == convertToInteger(t)) && (convertToInteger(aout[int(vi+1)]) == convertToInteger(0)) || a{	// dout - внеочередной тест {
				if (convertToInteger(math.Abs(float64(pol-man))) < convertToInteger(8)) {
					aout[int(vi+1)]=GETTICKS(0)	// ждем
				} else {
					dout[int(vi)]=1
				}
				dout[int(vi+3)]=0			// сбросить флаг внеочередного теста
			}

			if convertToInteger(aout[int(vi+1)]) != convertToInteger(0) {

				u=int(math.Min(float64(man)+15, 100))      // все время теста держим задание

				if (convertToInteger(GETTICKS(aout[int(vi+1)])*TICKSIZE()) >= convertToInteger(40)) || (convertToInteger(pol) > convertToInteger(math.Min(99, float64(man+8)))) {

					if (convertToInteger(pol) < convertToInteger(man+8)) {dout[int(vi)]=1} else {dout[int(vi)]= 0}
					aout[int(vi+1)]=0			// сам приедет обратно
				}
			}
		}

		aout[int(vi+2)]=h
	}
	return(u)
}

// переходы в режим по кнопкам или командам
// vi_mode - номер перем режима грс (уст извне, 0-по месту, 1-пу, 2-арм)
// evt - команда/кнопка
// vi  - номер переменной слежения
// v1,v2 - значения режима грс, между которыми переход
//
func hev(vi_mode int, vi int, evt *float32, v1 int, v2 int) {
	if (convertToInteger(front(evt,vi)) == 1) {		// нажата/подана {
		if (convertToInteger(dout[vi_mode]) == convertToInteger(v1)) {
			dout[vi_mode]=v2			// туды
		} else {
			if (convertToInteger(dout[vi_mode]) == convertToInteger(v2)) {
				dout[vi_mode]=v1       		// сюды
			}
		}
	}
}

//
// реакция на команды и кнопки перехода в режим
// по кнопкам или командам 1 переходы 0-2-0
// по кнопкам или командам 2 переходы 1-2-1
// переходы 0-1-0 запрещены
// cmd1,cmd2 - команды пользователя (упр извне вычислитель)
// but1,but2 - кнопки смены режима (д.вх)
// vi - начальный номер области переменных
// vi+0 - номер перем режима грс (уст извне, 0-по месту, 1-пу, 2-арм)
// vi+1..vi+4 - слежение за ком-кнопками
// vi+5, vi+6 - задержка при восст команд реж ту грс
// vi+7, vi+8 - тела команд реж ту грс
//
func modes(vi int, cmd1 *float32, cmd2 *float32, but1 *float32, but2 *float32) {
	hev(vi+0,vi+1,cmd1,0,2)
	hev(vi+0,vi+2,but1,0,2)
	hev(vi+0,vi+3,cmd2,1,2)
	hev(vi+0,vi+4,but2,1,2)
	if (convertToInteger(valTrack(cmd1, 5,float32(vi+5))) == 1) {
		dout[vi+7]=0
	}
	if (convertToInteger(valTrack(cmd2, 5,float32(vi+6))) == 1) {
		dout[vi+8]=0
	}
	//return(false)
}

// переходы в режим по команде от алгоритма
// vi_mode - номер перем режима грс (уст извне, 0-по месту, 1-пу, 2-арм)
// cmd - команда, значение режима грс, куда надо перевести
// vi  - номер переменной слежения
// vi+1 - тело команды
//
func cmdmode_in(vi_mode int, vi int, cmd *float32) {
	if DOST(cmd) && convertToInteger(cmd) != convertToInteger(dout[vi]) {
		dout[vi_mode]=int(*cmd)
		dout[vi+1]=dout[vi_mode]
	}
	dout[vi]=int(*cmd)
	dout[vi+1]=dout[vi_mode]
	//return(false)
}


func to_mest(vi_mode int, vi int) any {
	if (convertToInteger(dout[vi]) == convertToInteger(1)) {
		dout[vi_mode]=0	// по месту
		time.Sleep((3*18) * time.Second)
		dout[vi]=0		// взвод
	}
	return(false)
}

// Библиотека для работы с одоризаторами БОМ
// #include "eval.lib\BOM.evl"

// vi    - номер свободной выходной переменной для таймеров  Требуется 2 шт.
// q1    - мгновенный расход газа по прибору учета газа
// qz    - расход газа замещающий, при недостоверных
//	  	  данных с прибора учета газа (суперфло)
// mode  - режим одоризации
//	  	  для БОМ:  	0-автоматич от суперфло (посредством реле)
//		    		1-автоматич от САУ (требуется засылка)
//		    		2-ручное задание расхода газа
// cnt_sys - #сист номер расхода газа одоризатора в который засылать
// Т 	- период засылки.
// ПРИМЕР ИСПОЛЬЗОВАНИЯ:
// в ините
//  aout[16]=GETTICKS(0)
//  aout[17]=GETTICKS(0)
// в тексте
//  x=setq_periodic(16,(Reps["МГН РАС-1 ДЮРТ"].Value+Reps["МГН РАС-2 ДЮРТ"].Value),Reps["QМГН ЗАМ ДЮРТ"].Value,Reps["РЕЖ ОДОР ДЮРТ"].Value,Reps["QГ ОДОР ДЮРТ"].sys_num,30)
// или
//  x=setq_one((Reps["ПР СУТ-1 ДЮРТ"].Value+Reps["ПР СУТ-2 ДЮРТ"].Value)/24,Reps["QМГН ЗАМ ДЮРТ"].Value,Reps["QГ ОДОР ДЮРТ"].sys_num)


func setq(q float32, cnt_sys *float32) {
	if DOST(cnt_sys) {
		if (convertToInteger(math.Abs(float64(*cnt_sys-q))) > convertToInteger(5)) {  	// если расход мало изменился не засылаем {
			SET(cnt_sys, q)

		}
	}
}

func setq_one(q float32, qz float32, cnt_sys *float32) {
	if DOST(q) {
		setq(q,cnt_sys)
	} else {
		setq(qz,cnt_sys)
	}
}

func setq_periodic(vi int, q1 float32, qz float32, mode any, cnt_sys *float32, T any) any {
	if convertToInteger(mode) != convertToInteger(2) { // не автомат расход {
		return(false)
	}

	if (convertToInteger(GETTICKS(aout[vi])*TICKSIZE()) >= convertToInteger(T)) {
		if DOST(q1) {
			setq(q1,cnt_sys)
		} else {
			if valTrack(!DOST(q1),60,float32(vi+1)) {// если расход недост ждем 60 сек потом засылаем замещенный {
				setq(qz,cnt_sys)
			}
		}
		aout[vi]=GETTICKS(0)
	}

	return(false)
}