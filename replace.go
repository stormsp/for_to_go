package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

var (
	startTime      = time.Now()
	ticksPerSecond = 1000
)

func replaceAllStringRegexp(input, pattern, replace string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllString(input, replace)
}

func translate_for_to_go(code string) string {

	// Перевод символов
	code = strings.ReplaceAll(code, ";", "//")
	code = strings.ReplaceAll(code, "&", " && ")
	code = replaceAllStringRegexp(code, `(?i)\s*end\w*`, "\n}")
	code = replaceAllStringRegexp(code, `(?i)(func[ \t]+)(\w+\s*\([^)]*\))\s*`, "func $2 { \n")

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
	code = replaceAllStringRegexp(code, `(?i)\b(eq)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `($2 == $3)`)
	code = replaceAllStringRegexp(code, `(?i)\bne\(([^,]+?(?:\([^)]+\))?),([^)]+)\)`, `$1 != $2`)
	code = replaceAllStringRegexp(code, `(?i)\b(ge)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `($2 >= $3)`)
	code = replaceAllStringRegexp(code, `(?i)\b(lt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `($2 < $3)`)
	code = replaceAllStringRegexp(code, `(?i)\b(gt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `($2 > $3)`)
	code = replaceAllStringRegexp(code, `(?i)\b(le)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `($2 <= $3)`)
	code = replaceAllStringRegexp(code, `(?i)\b(NOT)\(([^)]+)\)`, `^(0xFFFFFFFFFFFFFFFF & $3)`)

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
	code = replaceAllStringRegexp(code, `(?i)\bset\s+([^,]+(?:\{[^}]+\})?),([^)]+)\s*\)?`, "SET($1, $2)\n")

	//dost
	code = replaceAllStringRegexp(code, `(?i)dost`, "DOST")

	//sleep
	code = replaceAllStringRegexp(code, `sleep\(([^)]+)\)`, `time.Sleep(($1) * time.Second)`)

	return code
}

func main() {
	code := `package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
	"time"
)
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
	codeFinal := code + outputCode

	// Запись данных в файл "output.go"
	err = ioutil.WriteFile("output.go", []byte(codeFinal), 0644)
	if err != nil {
		fmt.Println("Ошибка записи в файл:", err)
		return
	}

	fmt.Println("Операции чтения из файла и записи в файл успешно выполнены.")

}