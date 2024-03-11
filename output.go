package main

import (
	"time"
)

var aout [100]int
var dout [100]int
var x int

// Аварийный останов
// отличное от 0 не прошло timeout секунд.
// В противном случае функция возвращает 1.
func valTrack(val any, timeout int, id int) any {
	if val == 0 {
		aout[id] = 0
		return (0)
	}
	// aout[id] время перехода в состояние, отличное от 0
	// для вычисления тайм-аута (в тиках со старта зонда)
	if aout[id] == 0 {
		aout[id] = GETTICKS(0)
	}
	if GETTICKS(aout[id])*TICKSIZE() >= timeout {
		return (1)
	}
	return (0)
}

// setwex - аналог встроенной SET_WAIT
// однако, в случае не успеха
// производится дополнительные 2 попытки
// достигнуть заданного сотояния
func SET_ER03(SYS any, VAL any) any {
	if DOST(SYS) == 1 { // иначе нет связи с модулем
		SET(SYS, VAL)

	}
	return (0)
}

func setwex(sys any, state any, timeout any) any {
	if SET_WAIT(sys, state, timeout) != 0 {
		time.Sleep((18) * time.Second)
		return (SET_WAIT(sys, state, timeout))
	}
	return (0)
}
func setex(sys any, value any) any {
	if DOST(sys) == 0 {
		return (0)
	}
	SET(sys, value)
	return (1)
}

// ту при условии достоверности
func setwex1(sys any, value any) any {
	if DOST(sys) == 0 {
		return (0)
	}
	return (SET_WAIT(sys, value, 5))
}

// checkPrecond возвращает не ноль
// если выявлены условия
// работы алгоритма
func checkPrecondSt(dummy any) any {
	x = 0
	x = x | Reps["пожар ГРС ДЕС"].Value | Reps["1П СГ ЗАГ2 ДЕС"].Value | Reps["2П СГ ЗАГ2 ДЕС"].Value | Reps["3П СГ ЗАГ2 ДЕС"].Value | Reps["4П СГ ЗАГ2 ДЕС"].Value
	x = x | Reps["1Т СГ ЗАГ2 ДЕС"].Value | Reps["2Т СГ ЗАГ2 ДЕС"].Value | Reps["3Т СГ ЗАГ2 ДЕС"].Value | Reps["1О СГ ЗАГ2 ДЕС"].Value | Reps["2О СГ ЗАГ2 ДЕС"].Value
	x = x | Reps["1К СГ ЗАГ2 ДЕС"].Value | Reps["2К СГ ЗАГ2 ДЕС"].Value | Reps["КНОП АО ДЕС"].Value | Reps["АварЗакГРС ДЕС"].Value
	return (x)
}
func checkPrecondBt(dummy any) any {
	x = 0
	x = x | Reps["РвыхВР АС ДЕС"].Value | Reps["Рвых НР ДЕС"].Value | Reps["Кноп АО ДЕС"].Value
	x = x | Reps["ЗакГРСбСТР ДЕС"].Value | Reps["пад РвхГРС ДЕС"].Value | Reps["Рвх НР ДЕС"].Value
	x = x | Reps["ОШИБ БП ДЕС"].Value | (Reps["Рвых байп ДЕС"].Value >= Reps["ЗадPгВыхРабДЕС"].Value*1.15)
	return (x)
}

func oninit(t any) any {
	dout[1] = 0 // ход АО ст
	dout[2] = 0 // ход АО бс
	aout[3] = 0
	aout[4] = 0
	aout[5] = 0
	aout[6] = 0
	aout[7] = 0
	time.Sleep((5 * 18) * time.Second) // ждем первого опроса модулей
}

func main() {
	// Аварийное Закрытие ГРС со стравливанием
	if checkPrecondSt(0) {
		if Reps["РЕЖИМ ГРС ДЕС"].Value != 0 && (Reps["ХОД АО СТ ДЕС"].Value == 0) && (Reps["ХОД АО СТ ДЕС"].Value == 0) && (Reps["РЗР АО СТ ДЕС"].Value == 0) {
			dout[1] = 1 // ход ао
			// пошел останов
			SET(Reps["табл АварияДЕС"].sys_num, 1)
			SET(Reps["звонок шк ДЕС"].sys_num, 1)
			x = setwex(Reps["ОХР КР ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["Вход ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["Выход ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["ВЫХ Д ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["КРдоРУ ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			SET(Reps["Клап котлы ДЕС"].sys_num, 1)

			//set({}[sys_num],1 - отключение котлов
			if (Reps["Вход ДЕС"].Value == 2) && (Reps["Выход ДЕС"].Value == 2) && (Reps["ОХР КР ДЕС"].Value == 2) && (Reps["КРдоРУ ДЕС"].Value == 2) {
				x = setwex(Reps["СВзаВХ ДЕС"].sys_num, 0, Reps["Т ож кран ДЕС"].Value)  // открыть свечной кран на входе
				x = setwex(Reps["СВдоВЫХ ДЕС"].sys_num, 0, Reps["Т ож кран ДЕС"].Value) // открыть свечной кран на выходе
				x = setwex(Reps["СВ ОК ДЕС"].sys_num, 0, Reps["Т ож кран ДЕС"].Value)   // открыть свечной кран ОХРАН крана
				SET(Reps["ТУ КОТЕЛ 1 ДЕС"].sys_num, 1)
				// выключить котел 1
				SET(Reps["ТУ КОТЕЛ 2 ДЕС"].sys_num, 1)
				// выключить котел 2
			}
			dout[1] = 2 + ((Reps["ОХР КР ДЕС"].Value == 2) && (Reps["Вход ДЕС"].Value == 2) && (Reps["КРдоРУ ДЕС"].Value == 2) && (Reps["Выход ДЕС"].Value == 2) && (Reps["ВЫХ Д ДЕС"].Value == 2) && (Reps["СВзаВХ ДЕС"].Value == 1) && (Reps["СВдоВЫХ ДЕС"].Value == 1) && (Reps["СВ ОК ДЕС"].Value == 1)) // ход ао
			SET(Reps["КОМ РЕЖ3 ДЕС"].sys_num, 0)
			// перевод в информ режим
		}
	}
	// Закрытие ГРС без страваливания
	if checkPrecondBt(0) {
		if Reps["РЕЖИМ ГРС ДЕС"].Value != 0 && (Reps["ХОД АО БС ДЕС"].Value == 0) && (Reps["ХОД АО СТ ДЕС"].Value == 0) && (Reps["РЗР АО БС ДЕС"].Value == 0) {
			dout[2] = 1 // ход ао
			// пошел останов
			SET(Reps["табл АварияДЕС"].sys_num, 1)
			SET(Reps["звонок шк ДЕС"].sys_num, 1)
			x = setwex(Reps["ОХР КР ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["Вход ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["Выход ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["ВЫХ Д ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			x = setwex(Reps["КРдоРУ ДЕС"].sys_num, 1, Reps["Т ож кран ДЕС"].Value)
			SET(Reps["Клап котлы ДЕС"].sys_num, 1)
			if (Reps["Вход ДЕС"].Value == 2) && (Reps["Выход ДЕС"].Value == 2) && (Reps["ОХР КР ДЕС"].Value == 2) && (Reps["КРдоРУ ДЕС"].Value == 2) { // проверка закрылись ли вх и вых краны
			}
			dout[2] = 2 + ((Reps["ОХР КР ДЕС"].Value == 2) && (Reps["Вход ДЕС"].Value == 2) && (Reps["КРдоРУ ДЕС"].Value == 2) && (Reps["Выход ДЕС"].Value == 2) && (Reps["ВЫХ Д ДЕС"].Value == 2)) // ход ао
			SET(Reps["КОМ РЕЖ3 ДЕС"].sys_num, 0)
			// перевод в информ режим
		}
	}
	//
}
