package main

import (
	"fmt"
	"os/exec"
	"time"
)

// TODO: nmin, nmax,dost, true,false, cyclesec,execsec,
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

func SET_WAIT(parameter any, value any, timeout any) any {
	parameter = value
	return 0
}

func main() {
	dw := uint32(0b101)            // Пример двойного слова
	bitValue := BIT(dw, 0)         // Получение значения бита
	bitsValue := BITS(dw, 1, 0b11) // Получение группы битов
	// bxchgValue := BXCHG(dw, "1234") // Перестановка байтов, требуется реализация
	setbitsValue := SETBITS(0b1001, 2, 1, 0b11) // Установка битов

	fmt.Printf("BIT: %d\n", bitValue)
	fmt.Printf("BITS: %d\n", bitsValue)
	// fmt.Printf("BXCHG: %x\n", bxchgValue)
	fmt.Printf("SETBITS: %b\n", setbitsValue)
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
