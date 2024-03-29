package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"
)

var (
	startTime      = time.Now()
	ticksPerSecond = 1000
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
	code = strings.ReplaceAll(code, ";", "//")
	code = strings.ReplaceAll(code, "&", " && ")
	code = strings.ReplaceAll(code, "|", " || ")
	code = replaceAllStringRegexp(code, `#\[(.*?)\]`, "$1")
	code = replaceAllStringRegexp(code, `(?i)\s*end\w*`, "\n}")

	// изменение func и добавление any после каждой переменной
	code = replaceAllStringRegexpFunc(code, `(?i)(func[ \t]+)(\w+\s*\(\s*[^)]*\s*\))\s*`, func(match string) string {
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
	code = replaceAllStringRegexp(code, `func\s+(\w+)\s*\(([^)]*)\)`, `func $1($2`)

	//условия
	code = replaceAllStringRegexp(code, `(?i)IF\s*\((.+)\)`, `if $1 {`)
	code = replaceAllStringRegexp(code, `(?i)ELSE`, "} else {")

	// Перевод функций математики
	code = replaceAllStringRegexp(code, `(?i)abs\(([^)]+)\)`, `math.Abs($1)`)
	code = replaceAllStringRegexp(code, `(?i)acos\(([^)]+)\)`, `math.Acos($1)`)
	code = replaceAllStringRegexp(code, `(?i)asin\(([^)]+)\)`, `math.Asin($1)`)
	code = replaceAllStringRegexp(code, `(?i)atan\(([^)]+)\)`, `math.Atan($1)`)
	code = replaceAllStringRegexp(code, `(?i)cos\(([^)]+)\)`, `math.Cos($1)`)
	code = replaceAllStringRegexp(code, `(?i)sin\(([^)]+)\)`, `math.Sin($1)`)
	code = replaceAllStringRegexp(code, `(?i)tan\(([^)]+)\)`, `math.Tan($1)`)
	code = replaceAllStringRegexp(code, `(?i)exp\(([^)]+)\)`, `math.Exp($1)`)
	code = replaceAllStringRegexp(code, `(?i)ln\(([^)]+)\)`, `math.Log($1)`)
	code = replaceAllStringRegexp(code, `(?i)log\(([^)]+)\)`, `math.Log10($1)`)
	code = replaceAllStringRegexp(code, `(?i)sqrt\(([^)]+)\)`, `math.Sqrt($1)`)
	code = replaceAllStringRegexp(code, `(?i)sign\(([^)]+)\)`, `math.Sign($1)`)
	code = replaceAllStringRegexp(code, `(?i)sgn\(([^)]+)\)`, `(0 if $1 == 0 else -1 if $1 < 0 else 1)`)
	code = replaceAllStringRegexp(code, `(?i)pow\(([^,]+),([^)]+)\)`, `math.Pow($1, $2)`)
	code = replaceAllStringRegexp(code, `(?i)rand\(\)`, `random.Float64()`)
	code = replaceAllStringRegexp(code, `(?i)min\(([^,]+),([^)]+)\)`, `math.Min($1, $2)`)
	code = replaceAllStringRegexp(code, `(?i)max\(([^,]+),([^)]+)\)`, `math.Max($1, $2)`)
	code = replaceAllStringRegexp(code, `(?i)int\(([^)]+)\)`, `int($1)`)
	code = replaceAllStringRegexp(code, `(?i)restdiv\(([^,]+),([^)]+)\)`, `$1 % $2`)
	// Для функций nmin и nmax нужно будет написать функцию внутри Go
	code = replaceAllStringRegexp(code, `(?i)nmin\(([^)]+)\)`, `NMin($1)`)
	code = replaceAllStringRegexp(code, `(?i)nmax\(([^)]+)\)`, `NMax($1)`)

	// Логические операции
	// Dost TRUE FALSE нужно будет написать функции внутри GO, потому что такой альтернативы нет
	code = replaceAllStringRegexp(code, `(?i)\b(eq)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) == convertToInteger($3))`)
	code = replaceAllStringRegexp(code, `(?i)\bne\(([^,]+?(?:\([^)]+\))?),([^)]+)\)`, `convertToInteger($1) != convertToInteger($2)`)
	code = replaceAllStringRegexp(code, `(?i)\b(ge)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) >= convertToInteger($3))`)
	code = replaceAllStringRegexp(code, `(?i)\b(lt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) < convertToInteger($3))`)
	code = replaceAllStringRegexp(code, `(?i)\b(gt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) > convertToInteger($3))`)
	code = replaceAllStringRegexp(code, `(?i)\b(le)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(convertToInteger($2) <= convertToInteger($3))`)
	code = replaceAllStringRegexp(code, `(?i)\b(NOT)\(([^)]+)\)`, `^(0xFFFFFFFFFFFFFFFF & $3)`)

	//конвертируем в инт
	// Работа с битами и байтами
	// Функция BIT
	code = replaceAllStringRegexp(code, `bit\(([^,]+),([^)]+)\)`, `$1 & (1 << $2)`)
	// Функция BITS - может потребоваться вспомогательная функция
	code = replaceAllStringRegexp(code, `bits\(([^,]+),([^,]+),([^)]+)\)`, `Bits($1, $2, $3)`)
	// Функция BXCHG - может потребоваться вспомогательная функция
	code = replaceAllStringRegexp(code, `bxchg\(([^,]+),([^)]+)\)`, `Bxchg($1, $2)`)
	// Функция SETBITS - может потребоваться вспомогательная функция
	code = replaceAllStringRegexp(code, `setbits\(([^,]+),([^,]+),([^,]+),([^)]+)\)`, `Setbits($1, $2, $3, $4)`)

	// Функции над временем
	// Функция TIME
	code = replaceAllStringRegexp(code, `time\(\)`, `time.Now().Unix()`)
	// Функция SECOND
	code = replaceAllStringRegexp(code, `second\(([^)]+)\)`, `$1.Second()`)
	// Функция MINUTE
	code = replaceAllStringRegexp(code, `minute\(([^)]+)\)`, `$1.Minute()`)
	// Функция HOUR
	code = replaceAllStringRegexp(code, `hour\(([^)]+)\)`, `$1.Hour()`)
	// Функция MONTHDAY
	code = replaceAllStringRegexp(code, `monthday\(([^)]+)\)`, `$1.Day()`)
	// Функция MONTH
	code = replaceAllStringRegexp(code, `month\(([^)]+)\)`, `int($1.Month()) - 1`) // В Go месяцы начинаются с 1, а не с 0
	// Функция YEAR
	code = replaceAllStringRegexp(code, `year\(([^)]+)\)`, `$1.Year()`)
	// Функция WEEKDAY
	code = replaceAllStringRegexp(code, `weekday\(([^)]+)\)`, `int($1.Weekday())`) // В Go воскресенье - это 0
	// Функция YEARDAY
	code = replaceAllStringRegexp(code, `yearday\(([^)]+)\)`, `$1.YearDay() - 1`) // В Go дни года начинаются с 1
	// Функция MAKETIME
	code = replaceAllStringRegexp(code, `maketime\(([^,]+),([^,]+),([^,]+),([^,]+),([^,]+),([^)]+)\)`,
		`time.Date($6 + 1, time.Month($5 + 1), $4, $1, $2, $3, 0, time.UTC).Unix()`)

	//getticks
	code = replaceAllStringRegexp(code, `(?i)getticks`, "GETTICKS")
	//ticksize
	code = replaceAllStringRegexp(code, `(?i)ticksize`, "TICKSIZE")
	//set
	code = replaceAllStringRegexp(code, `(?i)\bset\s+([^,]+(?:\{[^}]+\})?),\s*([^)\s]+)\s*(?:\)|\b)`, "SET($1, $2)\n")
	code = replaceAllStringRegexp(code, `(?i)\bset\s*{([^,]+(?:\{[^}]+\})?),\s*([^)\s]+)\s*(?:\)|\b)`, "SET({$1, $2)\n\t")

	//set_wait доделать!
	code = replaceAllStringRegexp(code, `(?i)set_wait`, "SET_WAIT")
	//return
	code = replaceAllStringRegexp(code, `(?i)return`, "return")
	//dost
	code = replaceAllStringRegexp(code, `(?i)dost`, "DOST")

	//sleep
	code = replaceAllStringRegexp(code, `sleep\(([^)]+)\)`, `time.Sleep(($1) * time.Second)`)

	//реперы
	Reps = findReps(code)
	code = replaceExpressions(code, Reps)
	code = replaceAllStringRegexp(code, `\.Value\[(.*?)\]`, `.$1`)

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
