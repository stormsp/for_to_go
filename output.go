package main

import (
	"time"
)

var aout [100]int
var dout [100]int
var x bool


	// Аварийный останов
// отличное от 0 не прошло timeout секунд.
// В противном случае функция возвращает 1.
//
func valTrack(val any, timeout int, id int) any {
	if (convertToInteger(val) == convertToInteger(0)) {
  aout[id]=0
  return(0)
}
 // aout[id] время перехода в состояние, отличное от 0
 // для вычисления тайм-аута (в тиках со старта зонда)
 if (convertToInteger(aout[id]) == convertToInteger(0)) {
  aout[id]=GETTICKS(0)
}
 if (convertToInteger(GETTICKS(aout[id])*TICKSIZE()) >= convertToInteger(timeout)) {
  return(1)
}
 return(0)
}
//
// setwex - аналог встроенной SET_WAIT
// однако, в случае не успеха
// производится дополнительные 2 попытки
// достигнуть заданного сотояния
func SET_ER03(SYS any, VAL any) any {
	if (convertToInteger(DOST(SYS)) == convertToInteger(1)) { // иначе нет связи с модулем
    SET(SYS, VAL)

}
return(0)
}

func setwex(sys any, state any, timeout any) bool {
	if convertToInteger(SET_WAIT(sys,state,timeout)) != convertToInteger(0) {
    time.Sleep((18) * time.Second)
    return(SET_WAIT(sys,state,timeout))
}
 return false
}
func setex(sys any, value any) any {
	if (convertToInteger(DOST(sys)) == convertToInteger(0)) {
  return(0)
}
 SET(sys, value)
return(1)
}
// ту при условии достоверности
//
func setwex1(sys any, value any) any {
	if (convertToInteger(DOST(sys)) == convertToInteger(0)) {
    return(0)
}
  return(SET_WAIT(sys,value,5))
}
//
// checkPrecond возвращает не ноль
// если выявлены условия
// работы алгоритма
//
func checkPrecondSt(dummy any) bool {
	//x = 0
  x=x || Reps["пожар ГРС ДЕС"].Value || Reps["1П СГ ЗАГ2 ДЕС"].Value || Reps["2П СГ ЗАГ2 ДЕС"].Value || Reps["3П СГ ЗАГ2 ДЕС"].Value || Reps["4П СГ ЗАГ2 ДЕС"].Value
  x=x || Reps["1Т СГ ЗАГ2 ДЕС"].Value || Reps["2Т СГ ЗАГ2 ДЕС"].Value || Reps["3Т СГ ЗАГ2 ДЕС"].Value || Reps["1О СГ ЗАГ2 ДЕС"].Value || Reps["2О СГ ЗАГ2 ДЕС"].Value
  x=x || Reps["1К СГ ЗАГ2 ДЕС"].Value || Reps["2К СГ ЗАГ2 ДЕС"].Value || Reps["КНОП АО ДЕС"].Value || Reps["АварЗакГРС ДЕС"].Value
 return(x)
}
func checkPrecondBt(dummy any) bool {
	//x = 0
  x=x || Reps["РвыхВР АС ДЕС"].Value || Reps["Рвых НР ДЕС"].Value || Reps["Кноп АО ДЕС"].Value
  x=x || Reps["ЗакГРСбСТР ДЕС"].Value || Reps["пад РвхГРС ДЕС"].Value || Reps["Рвх НР ДЕС"].Value
  x=x || Reps["ОШИБ БП ДЕС"].Value || (convertToInteger(Reps["Рвых байп ДЕС"].Value) >= convertToInteger(convertToInteger(Reps["ЗадPгВыхРабДЕС"].Value)*convertToInteger(1.15)))
 return(x)
}
//
func oninit(t any) any {
	dout[1]=0	      // ход АО ст
  dout[2]=0 	// ход АО бс
  aout[3]=0
  aout[4]=0
  aout[5]=0
  aout[6]=0
  aout[7]=0
  time.Sleep((5*18) * time.Second)	// ждем первого опроса модулей
 return (t)
}

func main() {
// Аварийное Закрытие ГРС со стравливанием
if checkPrecondSt(0) {
  if convertToInteger(Reps["РЕЖИМ ГРС ДЕС"].Value) != convertToInteger(0) && (convertToInteger(Reps["ХОД АО СТ ДЕС"].Value) == convertToInteger(0)) && (convertToInteger(Reps["ХОД АО СТ ДЕС"].Value) == convertToInteger(0)) && (convertToInteger(Reps["РЗР АО СТ ДЕС"].Value) == convertToInteger(0)) {
  dout[1]=1	// ход ао
  // пошел останов
    SET(Reps["табл АварияДЕС"].sys_num, 1)
	SET(Reps["звонок шк ДЕС"].sys_num, 1)
	x=setwex(Reps["ОХР КР ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["Вход ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["Выход ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["ВЫХ Д ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["КРдоРУ ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    SET(Reps["Клап котлы ДЕС"].sys_num, 1)
	
 //set({}[sys_num],1 - отключение котлов
    if (convertToInteger(Reps["Вход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["Выход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["ОХР КР ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["КРдоРУ ДЕС"].Value) == convertToInteger(2)) {
          x=setwex(Reps["СВзаВХ ДЕС"].sys_num,0,Reps["Т ож кран ДЕС"].Value)       // открыть свечной кран на входе
          x=setwex(Reps["СВдоВЫХ ДЕС"].sys_num,0,Reps["Т ож кран ДЕС"].Value)      // открыть свечной кран на выходе
          x=setwex(Reps["СВ ОК ДЕС"].sys_num,0,Reps["Т ож кран ДЕС"].Value)        // открыть свечной кран ОХРАН крана
          SET(Reps["ТУ КОТЕЛ 1 ДЕС"].sys_num, 1)
	             // выключить котел 1
          SET(Reps["ТУ КОТЕЛ 2 ДЕС"].sys_num, 1)
	             // выключить котел 2
}
dout[1]=convertToInteger(2)+convertToInteger((convertToInteger(Reps["ОХР КР ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["Вход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["КРдоРУ ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["Выход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["ВЫХ Д ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["СВзаВХ ДЕС"].Value) == convertToInteger(1)) && (convertToInteger(Reps["СВдоВЫХ ДЕС"].Value) == convertToInteger(1)) && (convertToInteger(Reps["СВ ОК ДЕС"].Value) == convertToInteger(1)))  // ход ао
    SET(Reps["КОМ РЕЖ3 ДЕС"].sys_num, 0)
   // перевод в информ режим
}
}
// Закрытие ГРС без страваливания
if checkPrecondBt(0) {
  if convertToInteger(Reps["РЕЖИМ ГРС ДЕС"].Value) != convertToInteger(0) && (convertToInteger(Reps["ХОД АО БС ДЕС"].Value) == convertToInteger(0)) && (convertToInteger(Reps["ХОД АО СТ ДЕС"].Value) == convertToInteger(0)) && (convertToInteger(Reps["РЗР АО БС ДЕС"].Value) == convertToInteger(0)) {
  dout[2]=1	// ход ао
  // пошел останов
    SET(Reps["табл АварияДЕС"].sys_num, 1)
	SET(Reps["звонок шк ДЕС"].sys_num, 1)
	x=setwex(Reps["ОХР КР ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["Вход ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["Выход ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["ВЫХ Д ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    x=setwex(Reps["КРдоРУ ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value)
    SET(Reps["Клап котлы ДЕС"].sys_num, 1)
	if (convertToInteger(Reps["Вход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["Выход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["ОХР КР ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["КРдоРУ ДЕС"].Value) == convertToInteger(2)) { // проверка закрылись ли вх и вых краны
}
    dout[2]=convertToInteger(2)+convertToInteger((convertToInteger(Reps["ОХР КР ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["Вход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["КРдоРУ ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["Выход ДЕС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["ВЫХ Д ДЕС"].Value) == convertToInteger(2)))  // ход ао
    
    SET(Reps["КОМ РЕЖ3 ДЕС"].sys_num, 0)
   // перевод в информ режим
}
}
//
}