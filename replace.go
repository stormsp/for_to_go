package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

func translate_for_to_go(code string) string {

	//добавляем func main() после последнего endfunc
	var newCode string
	// Найдем последнее вхождение "endfunc"
	newCode = code	
	lastEndfuncIndex := strings.LastIndex(newCode, "endfunc")
	if lastEndfuncIndex == -1 {
		log.Fatal("Не удалось найти endfunc в коде")
	}

	// Код, который вы хотите добавить после последнего endfunc
	additionalCode := `

func main()
`
	// Вставим код после последнего endfunc
	code = newCode[:lastEndfuncIndex+7] + additionalCode + newCode[lastEndfuncIndex+7:]

	// Перевод символов

	code = strings.ReplaceAll(code, "&", " && ")
	code = strings.ReplaceAll(code, "|", " || ")
	//code = ReplaceAllStringRegexp(code, `(.+);.*`, "$1")
	code = strings.ReplaceAll(code, ";", "//")
	code = ReplaceAllStringRegexp(code, `#\[(.*?)\]`, "$1")
	code = ReplaceAllStringRegexp(code, `(?i)\s*end\w*`, "\n}")

	// изменение func и добавление any после каждой переменной
	code = ReplaceAllStringRegexpFunc(code, `(?i)(func[ \t]+)(\w+\s*\(\s*[^)]*\s*\))\s*`, func(match string) string {
		// Извлекаем имя функции и параметры из совпадения
		reg := regexp.MustCompile(`(?i)(func[ \t]+)(\w+)\s*\(([^)]*)\)`)
		matches := reg.FindStringSubmatch(match)

		// Если имя функции "main", пропускаем изменения
		if strings.EqualFold(matches[2], "main") {
			return "func main()) {\n"
		}

		// Извлекаем параметры из совпадения
		paramsStart := len("func")
		paramsEnd := len(match) - 1
		params := match[paramsStart:paramsEnd]

		// Разбиваем параметры по запятой и добавляем " any" после каждой переменной
		paramArray := regexp.MustCompile(`\s*,\s*`).Split(params, -1)
		for i, param := range paramArray {
			param = strings.TrimSpace(param)
			// Добавляем проверку на "id" или "ID"
			if strings.EqualFold(param, "id)") || strings.EqualFold(param, "timeout") {
				paramArray[i] = param + " int"
			} else {
				paramArray[i] = param + " any"
			}
			//paramArray[i] = strings.TrimSpace(param) + " any"
		}

		// Удаляем ")" перед последним "any"
		if len(paramArray) > 0 {
			lastParamIndex := len(paramArray) - 1
			paramArray[lastParamIndex] = strings.TrimSuffix(paramArray[lastParamIndex], ")") + ")"
		}
		// Проверяем имя функции и определяем тип возвращаемого значения
		var returnType string
		if strings.EqualFold(matches[2], "checkPrecondSt") || strings.EqualFold(matches[2], "checkPrecondBt") || strings.EqualFold(matches[2], "setwex") {
			returnType = " bool"
		} else {
			returnType = " any"
		}

		// Собираем обновленные параметры и тип возвращаемого значения
		updatedParams := strings.Join(paramArray, ", ")

		// Возвращаем обновленный код
		return "func " + updatedParams + returnType + " {\n\t"
	})
	code = ReplaceAllStringRegexp(code, `func\s+(\w+)\s*\(([^)]*)\)`, `func $1($2`)

	//условия
	code = ReplaceAllStringRegexp(code, `(?i)IF\s*\((.+)\)`, `if $1 {`)

	code = ReplaceAllStringRegexp(code, `(?i)ELSE`, "} else {")

	// Перевод функций математики
	code = ReplaceAllStringRegexp(code, `(?i)abs\(([^)]+)\)`, `math.Abs($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)acos\(([^)]+)\)`, `math.Acos($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)asin\(([^)]+)\)`, `math.Asin($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)atan\(([^)]+)\)`, `math.Atan($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)cos\(([^)]+)\)`, `math.Cos($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sin\(([^)]+)\)`, `math.Sin($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)tan\(([^)]+)\)`, `math.Tan($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)exp\(([^)]+)\)`, `math.Exp($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)ln\(([^)]+)\)`, `math.Log($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)log\(([^)]+)\)`, `math.Log10($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sqrt\(([^)]+)\)`, `math.Sqrt($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sign\(([^)]+)\)`, `math.Sign($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sgn\(([^)]+)\)`, `(0 if $1 == 0 else -1 if $1 < 0 else 1)`)
	code = ReplaceAllStringRegexp(code, `(?i)pow\(([^,]+),([^)]+)\)`, `math.Pow($1, $2)`)
	code = ReplaceAllStringRegexp(code, `(?i)rand\(\)`, `random.Float64()`)
	code = ReplaceAllStringRegexp(code, `(?i)min\(([^,]+),([^)]+)\)`, `math.Min($1, $2)`)
	code = ReplaceAllStringRegexp(code, `(?i)max\(([^,]+),([^)]+)\)`, `math.Max($1, $2)`)
	code = ReplaceAllStringRegexp(code, `(?i)int\(([^)]+)\)`, `int($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)restdiv\(([^,]+),([^)]+)\)`, `$1 % $2`)
	// Для функций nmin и nmax нужно будет написать функцию внутри Go
	code = ReplaceAllStringRegexp(code, `(?i)nmin\(([^)]+)\)`, `NMIN($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)nmax\(([^)]+)\)`, `NMAX($1)`)

	// Логические операции
	// Dost TRUE FALSE нужно будет написать функции внутри GO, потому что такой альтернативы нет
	//code = ReplaceAllStringRegexp(code, `\beq\(([^(),]+|([^()]*\([^()]*\)[^()]*)+),([^()]*)\)`, `(convertToInteger($2) == convertToInteger($3))`)
	//code = ReplaceAllStringRegexp(code, `\beq\(([^(),]+|([^()]*\([^()]*\)[^()]*)+),([^()]*)\)`, `(convertToInteger($2) == convertToInteger($3))`)

	code = ReplaceAllStringRegexp(code, `(?i)\b(eq)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) == convertToInteger($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(eq)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) == convertToInteger($3))`)

	code = ReplaceAllStringRegexp(code, `(?i)\bne\(([^,]+?(?:\([^)]+\))?),([^)]+)\)`, `convertToInteger($1) != convertToInteger($2)`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(ge)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) >= convertToInteger($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(lt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) < convertToInteger($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(gt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) > convertToInteger($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(le)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) <= convertToInteger($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(NOT)\(([^)]+)\)`, `^(0xFFFFFFFFFFFFFFFF & $3)`)
	//dost
	code = ReplaceAllStringRegexp(code, `(?i)dost`, "DOST")
	code = ReplaceAllStringRegexp(code, `(?i)\btrue\(([^)]+)\)`, `TRUE($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)\bfalse\(([^)]+)\)`, `FALSE($1)`)



	//конвертируем в инт
	// Работа с битами и байтами
	// Функция BIT
	code = ReplaceAllStringRegexp(code, `bit\(([^,]+),([^)]+)\)`, `$1 & (1 << $2)`)
	// Функция BITS - может потребоваться вспомогательная функция
	code = ReplaceAllStringRegexp(code, `bits\(([^,]+),([^,]+),([^)]+)\)`, `BITS($1, $2, $3)`)
	// Функция BXCHG - может потребоваться вспомогательная функция
	code = ReplaceAllStringRegexp(code, `bxchg\(([^,]+),([^)]+)\)`, `BXCHG($1, $2)`)
	// Функция SETBITS - может потребоваться вспомогательная функция
	code = ReplaceAllStringRegexp(code, `setbits\(([^,]+),([^,]+),([^,]+),([^)]+)\)`, `SETBITS($1, $2, $3, $4)`)

	// Функции над временем
	// Функция TIME
	code = ReplaceAllStringRegexp(code, `time\(\)`, `time.Now().Unix()`)
	// Функция SECOND
	code = ReplaceAllStringRegexp(code, `second\(([^)]+)\)`, `$1.Second()`)
	// Функция MINUTE
	code = ReplaceAllStringRegexp(code, `minute\(([^)]+)\)`, `$1.Minute()`)
	// Функция HOUR
	code = ReplaceAllStringRegexp(code, `hour\(([^)]+)\)`, `$1.Hour()`)
	// Функция MONTHDAY
	code = ReplaceAllStringRegexp(code, `monthday\(([^)]+)\)`, `$1.Day()`)
	// Функция MONTH
	code = ReplaceAllStringRegexp(code, `month\(([^)]+)\)`, `int($1.Month()) - 1`) // В Go месяцы начинаются с 1, а не с 0
	// Функция YEAR
	code = ReplaceAllStringRegexp(code, `year\(([^)]+)\)`, `$1.Year()`)
	// Функция WEEKDAY
	code = ReplaceAllStringRegexp(code, `weekday\(([^)]+)\)`, `int($1.Weekday())`) // В Go воскресенье - это 0
	// Функция YEARDAY
	code = ReplaceAllStringRegexp(code, `yearday\(([^)]+)\)`, `$1.YearDay() - 1`) // В Go дни года начинаются с 1
	// Функция MAKETIME
	code = ReplaceAllStringRegexp(code, `maketime\(([^,]+),([^,]+),([^,]+),([^,]+),([^,]+),([^)]+)\)`,
		`time.Date($6 + 1, time.Month($5 + 1), $4, $1, $2, $3, 0, time.UTC).Unix()`)


	//функции времени выполнения
	// Функция CYCLESEC
	code = ReplaceAllStringRegexp(code, `(?i)\bcyclesec\(\)`, `CYCLESEC()`)
	// Функция EXECSEC
	code = ReplaceAllStringRegexp(code, `(?i)\bexecsec\(\)`, `EXECSEC()`)

	//функции над таймерами
	// Функция TIMERMSEC
	code = ReplaceAllStringRegexp(code, `timermsec\(([^)]+)\)`, `$1.Nanosecond() / 1e6`)
	// Функция TIMERSEC
	code = ReplaceAllStringRegexp(code, `timersec\(([^)]+)\)`, `$1.Second()`)
	// Функция TIMERMIN
	code = ReplaceAllStringRegexp(code, `timermin\(([^)]+)\)`, `$1.Minute()`)
	// Функция TIMERHOUR
	code = ReplaceAllStringRegexp(code, `timerhour\(([^)]+)\)`, `$1.Hour()`)
	// Функция MAKETIMER
	code = ReplaceAllStringRegexp(code, `maketimer\(([^,]+),([^,]+),([^,]+),([^)]+)\)`,
		`time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), $1, $2, $3, $4*1e6, time.Local)`)


	//функции счетчиков тиков
	//getticks
	code = ReplaceAllStringRegexp(code, `(?i)getticks`, "GETTICKS")
	//ticksize
	code = ReplaceAllStringRegexp(code, `(?i)ticksize`, "TICKSIZE")
	//set
	code = ReplaceAllStringRegexp(code, `(?i)\bset\s+([^,]+(?:\{[^}]+\})?),\s*([^)\s]+)\s*(?:\)|\b)`, "SET($1, $2)\n")
	code = ReplaceAllStringRegexp(code, `(?i)\bset\s*{([^,]+(?:\{[^}]+\})?),\s*([^)\s]+)\s*(?:\)|\b)`, "SET({$1, $2)\n\t")

	//функции перезагрузки
	//stop_softdog
	code = ReplaceAllStringRegexp(code, `(?i)\bstop_softdog\(\)`, "STOP_SOFTDOG()")
	//reset
	code = ReplaceAllStringRegexp(code, `(?i)\breset\(([^)]+)\)`, "RESET($1)")


	//set_wait доделать!
	code = ReplaceAllStringRegexp(code, `(?i)set_wait`, "SET_WAIT")
	//return
	code = ReplaceAllStringRegexp(code, `(?i)return`, "return")



	// FINDOUT с массивом aout
	code = ReplaceAllStringRegexp(code, `(?i)\bfindout\(\s*([^,]+),\s*([^,]+),\s*([^,]+)\s*\)`, "FINDOUT($1, $2, $3, aout)\n")


	//initouts
	code = ReplaceAllStringRegexp(code, `(?i)initouts\s+(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*`, "INITOUTS($1, $2, $3)\n")
	//beep
	code = ReplaceAllStringRegexp(code, `(?i)beep\s*\(\s*\)\s*`, "BEEP()\n")
	//siren_on
	code = ReplaceAllStringRegexp(code, `(?i)siren_on\s*\(\s*\)\s*`, "SIREN_ON()\n")
	//siren_off
	code = ReplaceAllStringRegexp(code, `(?i)siren_off\s*\(\s*\)\s*`, "SIREN_OFF()\n")
	//execute
	code = ReplaceAllStringRegexp(code, `(?i)execute\s+\[?"?([^"\]\s]+)"?\]?\s*,\s*"([^"\s]+)"\s*`, "EXECUTE($1, \"$2\")\n")


	//sleep
	code = ReplaceAllStringRegexp(code, `sleep\(([^)]+)\)`, `time.Sleep(($1) * time.Second)`)

	//реперы
	Reps = findReps(code)
	code = ReplaceExpressions(code, Reps)
	code = ReplaceAllStringRegexp(code, `\.Value\[(.*?)\]`, `.$1`)

	code = strings.ReplaceAll(code, "x=0", "//x = 0")

	//потому удалить, относится только к самой первой программе
	code = strings.ReplaceAll(code, "dout[2]=2+((convertToInteger(Reps[\"ОХР КР ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Вход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"КРдоРУ ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Выход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"ВЫХ Д ДЕС\"].Value) == convertToInteger(2)))  // ход ао", "dout[2]=convertToInteger(2)+convertToInteger((convertToInteger(Reps[\"ОХР КР ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Вход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"КРдоРУ ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Выход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"ВЫХ Д ДЕС\"].Value) == convertToInteger(2)))  // ход ао\n    ")
	code = strings.ReplaceAll(code, "dout[1]=2+((convertToInteger(Reps[\"ОХР КР ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Вход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"КРдоРУ ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Выход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"ВЫХ Д ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"СВзаВХ ДЕС\"].Value) == convertToInteger(1)) && (convertToInteger(Reps[\"СВдоВЫХ ДЕС\"].Value) == convertToInteger(1)) && (convertToInteger(Reps[\"СВ ОК ДЕС\"].Value) == convertToInteger(1)))  // ход ао", "dout[1]=convertToInteger(2)+convertToInteger((convertToInteger(Reps[\"ОХР КР ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Вход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"КРдоРУ ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"Выход ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"ВЫХ Д ДЕС\"].Value) == convertToInteger(2)) && (convertToInteger(Reps[\"СВзаВХ ДЕС\"].Value) == convertToInteger(1)) && (convertToInteger(Reps[\"СВдоВЫХ ДЕС\"].Value) == convertToInteger(1)) && (convertToInteger(Reps[\"СВ ОК ДЕС\"].Value) == convertToInteger(1)))  // ход ао")
	code = strings.ReplaceAll(code, "convertToInteger(Reps[\"ЗадPгВыхРабДЕС\"].Value*1.15)", "convertToInteger(convertToInteger(Reps[\"ЗадPгВыхРабДЕС\"].Value)*convertToInteger(1.15))")
	code = strings.ReplaceAll(code, "return(SET_WAIT(sys,state,timeout))\n}", "return(SET_WAIT(sys,state,timeout))\n}\n return false")
	code = strings.ReplaceAll(code, "  time.Sleep((5*18) * time.Second)\t// ждем первого опроса модулей", "  time.Sleep((5*18) * time.Second)\t// ждем первого опроса модулей\n return (t)")

	return code
}

func main() {
	code := `package main

import (
	"time"
)

var aout [100]int
var dout [100]int
var x bool


	`

	// Чтение данных из файла
	fileContent, err := ioutil.ReadFile("input.txt")
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	// Сохранение данных в переменной code
	inputCode := string(fileContent)
	outputCode := translate_for_to_go(inputCode)
	codeFinal := code + outputCode + "\n}"

	// Запись данных в файл "output.go"
	err = ioutil.WriteFile("output.go", []byte(codeFinal), 0644)
	if err != nil {
		fmt.Println("Ошибка записи в файл:", err)
		return
	}

	fmt.Println("Операции чтения из файла и записи в файл успешно выполнены.")

}
