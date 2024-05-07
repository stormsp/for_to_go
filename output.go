package main

import (
	"time"
)

var aout [100]int
var dout [100]int
var x bool


	// ГРС ДЕСНА
// 02.2022
// управление оборудованием
// valTrack возращает 0 если val равен 0
// или если с момента перехода val в состояние,
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
// управление при условии связи с модулем (для ЭР03)
// sys - системный номер параметра ЭР03
// val - значение
//
func set_er03(sys any, val any) any {
	if (convertToInteger(DOST(sys)) == convertToInteger(1)) {	// иначе нет связи с модулем
    SET(sys, val)

}
return(0)
}
//
//управление при связи с модулем синхронно
func set_er03s(SYS any, VAL any) any {
	if (convertToInteger(DOST(SYS)) == convertToInteger(1)) {
    R=SET_WAIT(SYS,VAL,10)
}
return(0)
}
//
func setex(sys any, value any) any {
	if (convertToInteger(DOST(sys)) == convertToInteger(0)) {
  return(0)
}
 SET(sys, value)
return(1)
}
// управление при условии достоверности
//
func setwex(sys any, value any, timeout any) bool {
	if (convertToInteger(DOST(sys)) == convertToInteger(0)) {
    return(0)
}
  return(SET_WAIT(sys,value,timeout))
}
// front 1->0
// src - дискр сигнал
// previ - номер переменной слежения
//
func front0(src any, previ any) any {
	//x = 0
  if DOST(src) && convertToInteger(src) != convertToInteger(dout[previ]) && (convertToInteger(src) == convertToInteger(0)) {
    x=1
}
  dout[previ]=src
  return(x)
}
// front 0->1
// src - дискр сигнал
// previ - номер переменной слежения
//
func front1(src any, previ any) any {
	//x = 0
  if DOST(src) && convertToInteger(src) != convertToInteger(dout[previ]) && (convertToInteger(src) == convertToInteger(1)) {
    x=1
}
  dout[previ]=src
  return(x)
}
// -------------------- Слив конденсата
// vis,niz - выс,низ уровень конд
// kr_sys - sys крана слива
// tsl - время слива
// n_rez - переменная результата слива (0-норма, 1-таймаут, 2-неиспр дву)
// n_rez+1 - переменная слежения
//
func sliv(vis any, niz any, kr_sys any, n_rez any, tsl any) any {
	if (convertToInteger(aout[n_rez+1]) == convertToInteger(0)) && (convertToInteger(vis) == convertToInteger(1)) {      // если не следим и уровень конд высокий
    x=setwex(kr_sys,0,Reps["Т ОЖ КРАН ДЕС"].Value)				// откр сброс auma
    aout[n_rez+1]=GETTICKS(0)			// начали следить
}
  if convertToInteger(aout[n_rez+1]) != convertToInteger(0) {			// следим
    tas=(convertToInteger(GETTICKS(aout[n_rez+1])*TICKSIZE()) >= convertToInteger(tsl))
    if (convertToInteger(niz) == convertToInteger(1)) || tas {				// ур конд низкий или истекло время
      x=setwex(kr_sys,1,Reps["Т ОЖ КРАН ДЕС"].Value)			// закр сброс auma
      if vis {
        dout[n_rez]=1
      } else {
        dout[n_rez]=tas				// результат слива
}
      aout[n_rez+1]=0                           // не следим
}
}
}
//-------------------- Включение
// znp - значение ТС загазованности  &&  ! пожар
// V_FB_SYS - сист номер ТС состояния вент
// V_CNTSYS - системный номер выхода ЭР03 на включение вентилятора
// vi - номер переменной слеж
// vi+1 - номер переменной ПС
//
func CHK_VENT_ON(znp any, V_FB_SYS any, V_CNTSYS any, vi any) any {
	if front1(znp,vi) {
   A=SET_ER03S(V_CNTSYS,1)
    SLEEP(5*18)
    if !V_FB_SYS {
      dout[vi+1]=1
}
}
return(0)
}
//-------------------- Выключение по пожару
// poz - тс пожар
// v_csys - #sys выключения вентилятора (имп)
// vi - переменная для front
//
func chk_vent_off_p(vi any, poz any, v_csys any) any {
	if front1(poz,vi) {
    x=setex(v_csys,0)
}
return(0)
}
//-------------------- Выключение по окончанию загазованности
// zag - загазов
// v_csys - #sys выключения вентилятора (имп)
// vi - переменная для front
//
func chk_vent_off_z(vi any, zag any, v_csys any) any {
	if front0(zag,vi) {
    x=setex(v_csys,0)
}
return(0)
}
// продление сигнала загазованнности при его отключении
// вент выдул - тс загаз упал, а мы сымитировали, что упал позже
// zag - текущий сигнал загазованности
// vi   - номер переменной слежения dout[vi]
// vi+1 - номер переменной счетчика aout[vi+1]
// T    - задержка в с
// возврат - продленный сигнал загазованности
//
func zaglong(zag any, vi any, T any) any {
	ret=zag
  if (convertToInteger(zag) == convertToInteger(0)) && (convertToInteger(dout[vi]) == convertToInteger(1)) {
    aout[vi+1]=GETTICKS(0)		// только упал, начинаем отсчет
    ret=1				      // и имитируем, что еще стоит
  } else {
    if convertToInteger(aout[vi+1]) != convertToInteger(0) {           // отсчитываем хвост
      if (convertToInteger(zag) == convertToInteger(1)) {    		// но опять возникла загаз
        aout[vi+1]=GETTICKS(0)      // сбросил, сшиваем, ret=1
}
      a=(GETTICKS(0)-aout[vi+1])*TICKSIZE()
      if (convertToInteger(a) > convertToInteger(T)) {  			// после пропажи прошло еще время, с
        aout[vi+1]=0			// перестали считать
        ret=0				// опустили сами
      } else {
        ret=1                       // держим
}
}
}
  dout[vi]=zag				// следим здесь, нельзя раньше выходить
return(ret)
}
//
// Выходные дискретные переменные:
func oninit(T any) any {
	INITOUTS(1, 0, 40)
time.Sleep((48) * time.Second)
}

func main() {
//____________________Включение табло загазованности и пожара
//
// два табло в ББП по СН4
if (convertToInteger(((Reps["1П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["3П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["3П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["4П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["4П СГ СОСТ ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   if (convertToInteger(((Reps["1П СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["1П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2П СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["2П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["3П СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["3П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["4П СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["4П СГ СОСТ ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   SET(Reps["т1ББПвключ ДЕС"].SYS_NUM, 2)
x=set_er03s(Reps["табл АварияДЕС"].sys_num,1)
   } else {
   SET(Reps["т1ББПвключ ДЕС"].SYS_NUM, 1)
x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
}
} else {
SET(Reps["т1ББПвключ ДЕС"].SYS_NUM, 0)

}
SET(Reps["т2ББПвключ ДЕС"].SYS_NUM, ({т1)
ББПвключ ДЕС})
//
// два табло в ББТ по СН4
if (convertToInteger(((Reps["1Т СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1Т СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2Т СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2Т СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["3Т СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["3Т СГ СОСТ ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   if (convertToInteger(((Reps["1Т СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["1Т СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2Т СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["2Т СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["3Т СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["3Т СГ СОСТ ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   SET(Reps["т1ББТвключ ДЕС"].SYS_NUM, 2)
x=set_er03s(Reps["табл АварияДЕС"].sys_num,1)
   } else {
   SET(Reps["т1ББТвключ ДЕС"].SYS_NUM, 1)
x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
}
} else {
SET(Reps["т1ББТвключ ДЕС"].SYS_NUM, 0)

}
SET(Reps["т2ББТвключ ДЕС"].SYS_NUM, ({т1)
ББТвключ ДЕС})
//
// два табло в котельной + по угарному газу
if (convertToInteger(((Reps["1К СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2К СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["1СС ЗАГ 1П ДЕС"].Value) && (convertToInteger(Reps["1СС ДЕС"].Value) == convertToInteger(0))) || ((Reps["1СС ЗАГ 1П ДЕС"].Value) && (convertToInteger(Reps["1СС ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   if (convertToInteger(((Reps["1К СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["1К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2К СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["2К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["1СС ЗАГ 2П ДЕС"].Value) && (convertToInteger(Reps["1СС ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   SET(Reps["т1КОТвключ ДЕС"].SYS_NUM, 2)
x=set_er03s(Reps["табл АварияДЕС"].sys_num,1)
   } else {
   SET(Reps["т1КОТвключ ДЕС"].SYS_NUM, 1)
x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
}
} else {
SET(Reps["т1КОТвключ ДЕС"].SYS_NUM, 0)

}
SET(Reps["т2КОТвключ ДЕС"].SYS_NUM, ({т1)
КОТвключ ДЕС})//
//
// табло СН4 в Одоризаторной
if (convertToInteger(((Reps["1О СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1О СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2О СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2О СГ СОСТ ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   if (convertToInteger(((Reps["1О СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["1О СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2О СГ ЗАГ2 ДЕС"].Value) && (convertToInteger(Reps["2О СГ СОСТ ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
   SET(Reps["тУзОдвключ ДЕС"].SYS_NUM, 2)
x=set_er03s(Reps["табл АварияДЕС"].sys_num,1)
   } else {
   SET(Reps["тУзОдвключ ДЕС"].SYS_NUM, 1)
x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
}
} else {
SET(Reps["тУзОдвключ ДЕС"].SYS_NUM, 0)

}
//
//______________________включение вентилятора______________________
//
if (convertToInteger(Reps["РЕЖИМ ГРС ДЕС"].Value) == convertToInteger(3))
// вкл по загазованности ПОРОГ 1 или по кнопке
    z1=((Reps["1П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["3П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["3П СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["4П СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["4П СГ СОСТ ДЕС"].Value) == convertToInteger(0)))
    A=CHK_VENT_ON((z1 || Reps["ВенББП кнопДЕС"].Value) && !Reps["Пожар ГРС ДЕС"].Value,Reps["ВенББП состДЕС"].sys_num,Reps["Вент БлПер ДЕС"].sys_num,1)

    z2=((Reps["1Т СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1Т СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2Т СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2Т СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["3Т СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["3Т СГ СОСТ ДЕС"].Value) == convertToInteger(0)))
    A=CHK_VENT_ON((z2 || Reps["ВенББТ кнопДЕС"].Value) && !Reps["Пожар ГРС ДЕС"].Value,Reps["ВенББТ состДЕС"].sys_num,Reps["Вент БлТехнДЕС"].sys_num,3)

    z3=((Reps["1К СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2К СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2К СГ СОСТ ДЕС"].Value) == convertToInteger(0)) || (Reps["1СС ЗАГ 1П ДЕС"].Value) && (convertToInteger(Reps["1СС ДЕС"].Value) == convertToInteger(0)))
    A=CHK_VENT_ON((z3 || Reps["ВенКот кнопДЕС"].Value) && !Reps["Пожар ГРС ДЕС"].Value,Reps["ВенКот состДЕС"].sys_num,Reps["Вент котельДЕС"].sys_num,5)

    z4=((Reps["1О СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1О СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2О СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2О СГ СОСТ ДЕС"].Value) == convertToInteger(0)))
    A=CHK_VENT_ON((z4 || Reps["ВенУзО кнопДЕС"].Value) && !Reps["Пожар ГРС ДЕС"].Value,Reps["ВенУзО состДЕС"].sys_num,Reps["Вент БлОдорДЕС"].sys_num,7)
//
// выкл по пожару
//
  x=chk_vent_off_p(9,Reps["Пожар ГРС ДЕС"].Value,Reps["Вент БлПер ДЕС"].sys_num)
  x=chk_vent_off_p(10,Reps["Пожар ГРС ДЕС"].Value,Reps["Вент БлТехнДЕС"].sys_num)
  x=chk_vent_off_p(11,Reps["Пожар ГРС ДЕС"].Value,Reps["Вент котельДЕС"].sys_num)
  x=chk_vent_off_p(12,Reps["Пожар ГРС ДЕС"].Value,Reps["Вент БлОдорДЕС"].sys_num)
//
// выкл по окончанию загазованности
//
  x=chk_vent_off_z(13,zaglong(z1,15,10),Reps["Вент БлПер ДЕС"].sys_num)
  x=chk_vent_off_z(16,zaglong(z2,17,60),Reps["Вент БлТехнДЕС"].sys_num)
  x=chk_vent_off_z(19,zaglong(z3,20,60),Reps["Вент котельДЕС"].sys_num)
  x=chk_vent_off_z(22,zaglong(z4,23,60),Reps["Вент БлОдорДЕС"].sys_num)
}
//
//____________________блокировка котлов отсека ПТН
//
if (convertToInteger(Reps["авар ОПТ ДЕС"].Value || Reps["прор ПГ3.1 ДЕС"].Value || Reps["прор ПГ3.2 ДЕС"].Value || ((Reps["1К СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["1К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["2К СГ ЗАГ1 ДЕС"].Value) && (convertToInteger(Reps["2К СГ СОСТ ДЕС"].Value) == convertToInteger(0))) || ((Reps["1СС ЗАГ 1П ДЕС"].Value) && (convertToInteger(Reps["1СС ДЕС"].Value) == convertToInteger(0))) || ((Reps["1СС ЗАГ 1П ДЕС"].Value) && (convertToInteger(Reps["1СС ДЕС"].Value) == convertToInteger(0)))) == convertToInteger(1))
  x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
  x=set_er03s(Reps["Клап котлы ДЕС"].sys_num,1)
  x=set_er03s(Reps["ТУ КОТЕЛ 1 ДЕС"].sys_num,1)
  x=set_er03s(Reps["ТУ КОТЕЛ 2 ДЕС"].sys_num,1)
  x=set_er03s(Reps["ТУ НАС К1 ДЕС"].sys_num,0)
  x=set_er03s(Reps["ТУ НАС К2 ДЕС"].sys_num,0)
  x=set_er03s(Reps["ТУ НАСПГ1 ДЕС"].sys_num,0)
  x=set_er03s(Reps["ТУ НАСПГ2 ДЕС"].sys_num,0)
  x=set_er03s(Reps["ТУ НАСОТ1 ДЕС"].sys_num,0)
  x=set_er03s(Reps["ТУ НАСОТ2 ДЕС"].sys_num,0)
  x=set_er03s(Reps["ТУ НАСПОДП ДЕС"].sys_num,0)
  dout[33]=1 //сигнал блок упр насосами
}
//
// -----------------Слив конденсата----------------------
//
// если разрешен слив
if (convertToInteger(Reps["РЕЖИМ ГРС ДЕС"].Value) == convertToInteger(3)) && (convertToInteger(Reps["ХОД АО СТ ДЕС"].Value) == convertToInteger(0)) && (convertToInteger(Reps["ХОД АО БС ДЕС"].Value) == convertToInteger(0)) && !Reps["РЗР СЛИВ ДЕС"].Value {
  if (convertToInteger(Reps["LкондФ1 ВР ДЕС"].Value) == convertToInteger(1)) // уровень высокий
  x=setwex(Reps["СлФС1 ДЕС"].sys_num,0,Reps["Т ож кран ДЕС"].Value) // открыть кран сброса
  sleep (50) // пауза
     if eq (Reps["LкондФ1 ВР ДЕС"].Value && Reps["LкондФ1 НР ДЕС"].Value,1) // проверка нет падения уровня
     SET(Reps["РЗР СЛИВ ДЕС"].sys_num, 1)
 // снятие авт режима слива
     x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
}
}
}
//
if (convertToInteger(Reps["LкондФ1 НР ДЕС"].Value) == convertToInteger(1)) // уровень низкий
  x=setwex(Reps["СлФС1 ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value) // закрыть кран сброса
}
//
if (convertToInteger(Reps["РЕЖИМ ГРС ДЕС"].Value) == convertToInteger(3)) && (convertToInteger(Reps["ХОД АО СТ ДЕС"].Value) == convertToInteger(0)) && (convertToInteger(Reps["ХОД АО БС ДЕС"].Value) == convertToInteger(0)) && !Reps["РЗР СЛИВ ДЕС"].Value {
  if (convertToInteger(Reps["LкондФ2 ВР ДЕС"].Value) == convertToInteger(1))
  x=setwex(Reps["СлФС2 ДЕС"].sys_num,0,Reps["Т ож кран ДЕС"].Value)
  sleep (50)
     if eq (Reps["LкондФ2 ВР ДЕС"].Value && Reps["LкондФ2 НР ДЕС"].Value,1)
     SET(Reps["РЗР СЛИВ ДЕС"].sys_num, 1)
 // снятие авт режима слива
     x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)
}
}
}
//
if (convertToInteger(Reps["LкондФ2 НР ДЕС"].Value) == convertToInteger(1))
  x=setwex(Reps["СлФС2 ДЕС"].sys_num,1,Reps["Т ож кран ДЕС"].Value) // закрыть кран сброса
}
//
//
//----------------Переключение насосов-------------------
//
if (convertToInteger(Reps["РЕЖИМ ГРС ДЕС"].Value) == convertToInteger(3)) && !Reps["РЗР насос ДЕС"].Value && !Reps["БЛОК НАС ДЕС"].Value
  if front1((convertToInteger(Reps["Время ДЕС"].Value.Hour()) == convertToInteger(Reps["ЧАС НАСОС ДЕС"].Value)),39)
      if !(Reps["ТС НАСПГ1 ДЕС"].Value) && (Reps["ТС НАСПГ2 ДЕС"].Value)  // насос1 в работе и насос2 отключен
      SET(Reps["ТУ НАСПГ2 ДЕС"].sys_num, 1)
         // включить насос 2
      SET(Reps["ТУ НАСПГ1 ДЕС"].sys_num, 0)
         // отключить насос 1
      } else {
        if !(Reps["ТС НАСПГ2 ДЕС"].Value) && (Reps["ТС НАСПГ1 ДЕС"].Value)   //насос2 в работе и насос1 отключен
        SET(Reps["ТУ НАСПГ1 ДЕС"].sys_num, 1)
         // включить насос 1
        SET(Reps["ТУ НАСПГ2 ДЕС"].sys_num, 0)
         // отключить насос 2
}
}
    if !(Reps["ТС НАСОТ1 ДЕС"].Value) && (Reps["ТС НАСОТ2 ДЕС"].Value)    // насос1 в работе и насос2 отключен
      SET(Reps["ТУ НАСОТ2 ДЕС"].sys_num, 1)
         // включить насос 2
      SET(Reps["ТУ НАСОТ1 ДЕС"].sys_num, 0)
         // отключить насос 1
    } else {
      if !(Reps["ТС НАСОТ2 ДЕС"].Value) && (Reps["ТС НАСОТ1 ДЕС"].Value)  //насос2 в работе и насос1 отключен
      SET(Reps["ТУ НАСОТ1 ДЕС"].sys_num, 1)
         // включить насос 1
      SET(Reps["ТУ НАСОТ2 ДЕС"].sys_num, 0)
         // отключить насос 2
}
}
}
}
//
//----------------управление подпиточным насосом-------------------
//г
if (convertToInteger(Reps["Ртн кон подДЕС"].Value) <= convertToInteger(0.24))                // давление в контуре подпитки мало
   SET(Reps["ТУ НАСПОДП ДЕС"].sys_num, 1)
          // включить насос
}
//
if (convertToInteger(Reps["Ртн кон подДЕС"].Value) >= convertToInteger(0.3))                 // давление в контуре подпитки выросло
   SET(Reps["ТУ НАСПОДП ДЕС"].sys_num, 0)
          // выключить насос
}
//
if (convertToInteger(Reps["НАСПОДП ДЕС"].Value) == convertToInteger(1))                      // при сигнале аварии насоса
   SET(Reps["ТУ НАСПОДП ДЕС"].sys_num, 0)
          // выключить насос
   x=set_er03s(Reps["звонок шк ДЕС"].sys_num,1)  // включить ПС
}


}