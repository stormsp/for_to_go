package main

import (
	"time"
	"math"
)

var aout [100]int
var dout [100]int
var x bool


	// ГРС Красноусольск (Стерлитамакское ЛПУМГ)
// 09.2019 Галеев
// Аварийный останов

// v2  -  Добавлены диагностические параметры для контроля выполнения алгоритма
// dout[1] - команда АО
// dout[2] - ход выполнения 0-нет , 1-выполняется
// dout[3] - причина сработки
// aout[4] - код ошибки выполнения
// aout[5] - дата последнего выполнения
// aout[6] - дата окончания выполнения


//valtrack.evl
// для вставки #include "eval.lib\ValTrack.evl"

//
// valTrack возращает 0 если val равен 0
// или если с момента перехода val в состояние,
// отличное от 0 не прошло timeout секунд.
// В противном случае функция возвращает 1.
//
func valTrack(val any, timeout int, id int) any {
	if (convertToInteger(val) == convertToInteger(0)) {
  aout[id]=0
  return(false)
}

 // aout[id] время перехода в состояние, отличное от 0
 // для вычисления тайм-аута (в тиках со старта зонда)
 if (convertToInteger(aout[id]) == convertToInteger(0)) {
  aout[id]=GETTICKS(0)
}

 if (convertToInteger(GETTICKS(aout[id])*TICKSIZE()) >= convertToInteger(timeout)) {
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
func valTrackGt(val any, bound any, timeout int, id int) any {
	return(valTrack(convertToInteger(DOST(val)) != convertToInteger(0) && (convertToInteger(val) > convertToInteger(bound)),timeout,id))
}

func valTrackLt(val any, bound any, timeout int, id int) any {
	return(valTrack(convertToInteger(DOST(val)) != convertToInteger(0) && (convertToInteger(val) < convertToInteger(bound)),timeout,id))
}

// при достоверности одного из трех каналов, недостоверный канал заменяется
// значением параметров с ЭКМ давление на вых низкое, высокое
//
func valTrackLt_DOST(val any, bound any, timeout int, id int, p_ekm any) any {
	if DOST(val) {
    return(valTrack((convertToInteger(val) < convertToInteger(bound)),timeout,id))
  } else {
    return(p_ekm)
}
}


func valTrackGt_DOST(val any, bound any, timeout int, id int, p_ekm any) any {
	if DOST(val) {
    return(valTrack((convertToInteger(val) > convertToInteger(bound)),timeout,id))
  } else {
    return(p_ekm)
}
}

//#include "eval.lib\set.evl"

// 01.06.15
// для вставки #include "eval.lib\set.evl"

//
// управление при условии достоверности
//
func setex(sys any, value any) bool {
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
func setwex(sys any, value any, timeout any) bool {
	if convertToInteger(SET_WAIT(sys,value,timeout)) != convertToInteger(0) {
    time.Sleep((18) * time.Second)
    return(SET_WAIT(sys,value,timeout))
}
  return(false)
}

//
// impuls
//
func impuls(sys any, t any) any {
	x=SET_WAIT(sys,1,t)
  time.Sleep((2*18) * time.Second)
  x=SET_WAIT(sys,0,t)
  return(x)
}



// установка значения с заданной чувствительностью
// возврат 1-установлено
//         0-без реакции
func setSens(sys int, value int, sens any) any {
	//x = 0
  if (convertToInteger(math.Abs(float64(sys - value))) > convertToInteger(sens)) {
    x=setex(sys,value)
}
  return(x)
}



func setwex_DOST(sys any, value any, timeout any) any {
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
func front(src int, id int) bool {
	//x = 0
  if DOST(src) && convertToInteger(src) != convertToInteger(dout[id]) && convertToInteger(src) != convertToInteger(0) {
    x=true
}
  dout[id]=src
  return(x)
}




//----------- Условия выполнения аварийного останова -----------------------------
  // по команде с экрана АРМ или от Диспетчера
  // При нажатой более 4 секунд кнопке(механической)- только при аварийной ситуации:
        // - Аварийно-высокое давление
        // - Пожар в операторной
        // - Пожар в блоке переключения (при наличии пож.сигнализации)
        // - Пожар в блоке одоризации (при наличии пож.сигнализации)
//-------------------------------------------------------------------------------
func checkFire(dummy any) bool {
	//x = 0
  x=x || Reps["ПОЖАР ОПЕ КРАС"].Value //<Пожар в операторной>.
  x=x || Reps["ПОЖАР ПЕР КРАС"].Value //<Пожар в блоке переключения>.
  //x=x || Reps["ПОЖАР ОДО КРАС"].Value //<Пожар в блоке одоризации>.
  return(x)
}


func checkPrecond(dummy any) any {
	var x bool
	//x = 0
  if convertToInteger(Reps["РЕЖИМ ГРС КРАС"].Value) != convertToInteger(0) {
    x=x||Reps["КОМ АО КРАС"].Value  //1 команда - без условий
    if (convertToInteger(valTrack(Reps["КН АВОСТ КРАС"].Value, 4, 8)) == convertToInteger(1)) {    // кнопка - только при аварийной ситуации {
      x=x||checkFire(false)        //2 Пожар
      x=x||Reps["РВЫХ123АВ КРАС"].Value    //3 Аварийно-высокое давление
}
}
  return(x)
}
//--------------------------------------------------------------------------------

func oninit(t any) any {
	dout[1]=0
 dout[2]=0
 dout[3]=0
 aout[4]=0
 aout[5]=1
 aout[6]=1

 // ждем первого опроса модулей
 time.Sleep((10*18) * time.Second)
 return nil
}

func main() {
reason:=checkPrecond(0)
if convertToInteger(reason) != convertToInteger(0) {

   dout[2]=1	// ход ао
   dout[3]=convertToInteger(reason)
   aout[5]=int(time.Now().Unix())

   // закрыть охранный кран
   x=setwex(Reps["КРАН ОХР КРАС"].sys_num,1,40)

   time.Sleep((18) * time.Second)
   if convertToInteger(Reps["КРАН ОХР КРАС"].Value) != convertToInteger(2) {
     // закрыть входной кран
     x=setwex(Reps["КРАН ВХОД КРАС"].sys_num,1,20)
}

   // закрыть байпасный кран
   x=setwex(Reps["КРАН БАЙП КРАС"].sys_num,1,20)

   // закрыть выходной
   x=setwex(Reps["КРАН ВЫХ КРАС"].sys_num,1,20)

   // подогреватель отключить
   x=SET_WAIT(Reps["ПГ УПР КРАС"].sys_num,2,20) 	

   // отключить одоризатор 
   x=SET_WAIT(Reps["РЕЖ ОДОР1 КРАС"].sys_num,0,20)

   // Если пожар
   if checkFire(0) {

     // если закрыты : Охранный, байпасный, выходной краны
     if (convertToInteger(Reps["КРАН ОХР КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["КРАН ВЫХ КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps["КРАН БАЙП КРАС"].Value) == convertToInteger(2)) {
       // открыть свечные краны
       x=setwex(Reps["КР СВ НИЗ КРАС"].sys_num,0,30)
       x=setwex(Reps["КР СВ ВЫС КРАС"].sys_num,0,30)
}

     // если охранный кран не закрыт, а закрыты: входной, байпасный, выходной краны
     if ((convertToInteger(Reps["КРАН ОХР КРАС"].Value) != convertToInteger(2)) && (convertToInteger(Reps["КРАН ВХОД КРАС"].Value) == convertToInteger(2))) && (convertToInteger(Reps["КРАН ВЫХ КРАС"].Value) == convertToInteger(2)) {
       // открыть свечной кран с низ стороны
       x=setwex(Reps["КР СВ НИЗ КРАС"].sys_num,0,30)
}
}
  
   // переводим грс в режим по месту
   SET(Reps["КОМ РЕЖ3"].sys_num, 0)
time.Sleep((5*18) * time.Second)
   dout[1]=0	// ком ао (возм причина)
   dout[2]=0

   aout[6]=int(time.Now().Unix())
}

if front(convertToInteger(convertToInteger(Reps["РЕЖИМ ГРС КРАС"].Value) != convertToInteger(0)),9) {
  dout[3]=0
}

}