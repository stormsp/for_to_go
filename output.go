package main

import (
	"time"
)

var aout [100]any
var dout [100]any
	// Аварийный останов
// отличное от 0 не прошло timeout секунд.
// В противном случае функция возвращает 1.
//
func valTrack(val any, timeout any, id any) any {
	if (val == 0) {
  aout[id]=0
  return(0)
}
 // aout[id] время перехода в состояние, отличное от 0
 // для вычисления тайм-аута (в тиках со старта зонда)
 if (aout[id] == 0) {
  aout[id]=GETTICKS(0)
}
 if (GETTICKS(aout[id])*TICKSIZE() >= timeout) {
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
	if (DOST(SYS) == 1) { // иначе нет связи с модулем
    SET SYS, VAL
}
return(0)
}

func setwex(sys any, state any, timeout any) any {
	if SET_WAIT(sys,state,timeout) != 0 {
    time.Sleep((18) * time.Second)
    return(SET_WAIT(sys,state,timeout))
}
  return(0)
}
func setex(sys any, value any) any {
	if (DOST(sys) == 0) {
  return(0)
}
 SET(sys, value)
return(1)
}
// ту при условии достоверности
//
func setwex1(sys any, value any) any {
	if (DOST(sys) == 0) {
    return(0)
}
  return(SET_WAIT(sys,value,5))
}
//
// checkPrecond возвращает не ноль
// если выявлены условия
// работы алгоритма
//
func checkPrecondSt(dummy any) any {
	x=0
  x=x|{пожар ГРС ДЕС}|{1П СГ ЗАГ2 ДЕС}|{2П СГ ЗАГ2 ДЕС}|{3П СГ ЗАГ2 ДЕС}|{4П СГ ЗАГ2 ДЕС}
  x=x|{1Т СГ ЗАГ2 ДЕС}|{2Т СГ ЗАГ2 ДЕС}|{3Т СГ ЗАГ2 ДЕС}|{1О СГ ЗАГ2 ДЕС}|{2О СГ ЗАГ2 ДЕС}
  x=x|{1К СГ ЗАГ2 ДЕС}|{2К СГ ЗАГ2 ДЕС}|{КНОП АО ДЕС}|{АварЗакГРС ДЕС}
 return(x)
}
func checkPrecondBt(dummy any) any {
	x=0
  x=x|{РвыхВР АС ДЕС}|{Рвых НР ДЕС}|{Кноп АО ДЕС}
  x=x|{ЗакГРСбСТР ДЕС}|{пад РвхГРС ДЕС}|{Рвх НР ДЕС}
  x=x|{ОШИБ БП ДЕС}|({Рвых байп ДЕС} >= {ЗадPгВыхРабДЕС}*1.15)
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
}
// Аварийное Закрытие ГРС со стравливанием
if checkPrecondSt(0) {
  if {РЕЖИМ ГРС ДЕС} != 0 && ({ХОД АО СТ ДЕС} == 0) && ({ХОД АО СТ ДЕС} == 0) && ({РЗР АО СТ ДЕС} == 0) {
  dout[1]=1	// ход ао
  // пошел останов
    set{табл АварияДЕС}[sys_num],1
    set{звонок шк ДЕС}[sys_num],1
    x=setwex({ОХР КР ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({Вход ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({Выход ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({ВЫХ Д ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({КРдоРУ ДЕС}[sys_num],1,{Т ож кран ДЕС})
    set{Клап котлы ДЕС}[sys_num],1
 //set({}[sys_num],1 - отключение котлов
    if ({Вход ДЕС} == 2) && ({Выход ДЕС} == 2) && ({ОХР КР ДЕС} == 2) && ({КРдоРУ ДЕС} == 2) // проверка закрылись ли краны
          x=setwex({СВзаВХ ДЕС}[sys_num],0,{Т ож кран ДЕС})       // открыть свечной кран на входе
          x=setwex({СВдоВЫХ ДЕС}[sys_num],0,{Т ож кран ДЕС})      // открыть свечной кран на выходе
          x=setwex({СВ ОК ДЕС}[sys_num],0,{Т ож кран ДЕС})        // открыть свечной кран ОХРАН крана
          set{ТУ КОТЕЛ 1 ДЕС}[sys_num],1             // выключить котел 1
          set{ТУ КОТЕЛ 2 ДЕС}[sys_num],1             // выключить котел 2
}
dout[1]=2+(({ОХР КР ДЕС} == 2) && ({Вход ДЕС} == 2) && ({КРдоРУ ДЕС} == 2) && ({Выход ДЕС} == 2) && ({ВЫХ Д ДЕС} == 2) && ({СВзаВХ ДЕС} == 1) && ({СВдоВЫХ ДЕС} == 1) && ({СВ ОК ДЕС} == 1))  // ход ао
    SET({КОМ РЕЖ3 ДЕС}[sys_num], 0)
   // перевод в информ режим
}
}
// Закрытие ГРС без страваливания
if checkPrecondBt(0) {
  if {РЕЖИМ ГРС ДЕС} != 0 && ({ХОД АО БС ДЕС} == 0) && ({ХОД АО СТ ДЕС} == 0) && ({РЗР АО БС ДЕС} == 0) {
  dout[2]=1	// ход ао
  // пошел останов
    set{табл АварияДЕС}[sys_num],1
    set{звонок шк ДЕС}[sys_num],1
    x=setwex({ОХР КР ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({Вход ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({Выход ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({ВЫХ Д ДЕС}[sys_num],1,{Т ож кран ДЕС})
    x=setwex({КРдоРУ ДЕС}[sys_num],1,{Т ож кран ДЕС})
    set{Клап котлы ДЕС}[sys_num],1
    if ({Вход ДЕС} == 2) && ({Выход ДЕС} == 2) && ({ОХР КР ДЕС} == 2) && ({КРдоРУ ДЕС} == 2) // проверка закрылись ли вх и вых краны
}
    dout[2]=2+(({ОХР КР ДЕС} == 2) && ({Вход ДЕС} == 2) && ({КРдоРУ ДЕС} == 2) && ({Выход ДЕС} == 2) && ({ВЫХ Д ДЕС} == 2))  // ход ао
    SET({КОМ РЕЖ3 ДЕС}[sys_num], 0)
   // перевод в информ режим
}
}
//