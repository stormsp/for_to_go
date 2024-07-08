package main

import (
	"fmt"
)

var dout [100]int

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

func main() {
	bp := 0
	cod1 := float32(202.5) // Пример значения
	cod2 := float32(170.7) // Пример значения

	bomOven(bp, cod1, cod2)

	// Выводим результат для проверки
	for i := 0; i < 19; i++ {
		fmt.Printf("dout[%d] = %d\n", bp+i, dout[bp+i])
	}
}