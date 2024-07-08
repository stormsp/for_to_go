package main

import (
	"time"
	"math"
)

var aout [100]int
var dout [100]int
var x bool


	// ГРС Красноусольский
// Галеев 07.2019
// Мелкие алгоритмы, влияющие на объекты

//
// час текущего времени сау
//
func curdummy any.Hour() any {
	curtime=time.Now().Unix()
 return(curtime.Hour())
}


#include "eval.lib\set.evl"
#include "eval.lib\valtrack.evl"

#include "eval.lib\front.evl"
#include "eval.lib\regim.evl"
#include "eval.lib\bom.evl"


// Переменные
// 1,2 - оператор 1,2 зарег
// 4 - режим ту грс
// 5-8 - слеж за ком-кнопками
// 9-10 - задержка при восст команд реж ту грс
// 11-12 - тела команд реж ту грс
// 13-14 - расчет связи c бусами
// 15 - расчет времени отсутствия связи с бусами
// 16 - слеж за временем одоризатор
// 18 - слеж за ком перевода в режим от алг
// 19 - тело команды перевода в режим от алгоритма
//
func oninit(t any) any {
	b=0
d=0
  prevhour=cur0.Hour()
  dout[1]=0
  dout[2]=0
  aout[3]=0
  dout[4]=Reps["АВТОЗАПУСКР"].Value          // уст извне, сохранение при перезагр
  dout[5]=0                      // при старте, считаем, кнопки не нажаты
  dout[6]=0
  dout[7]=0
  dout[8]=0
  aout[9]=0
  aout[10]=0
  dout[11]=0
  dout[12]=0
  aout[15]=0

  dout[18]=Reps["АВТОЗАПУСКР"].Value
  dout[19]=Reps["АВТОЗАПУСКР"].Value
  dout[21]=0
  aout[22]=0
  dout[23]=0

  aout[24]=0
  aout[25]=0
  aout[26]=0
  aout[27]=0
  time.Sleep((5*18) * time.Second)
}

func main() {
// реакция на команды и кнопки перехода в режим
//
x=modes(4,Reps["КОМ РЕЖ1 КРАС"].Value,Reps["КОМ РЕЖ2 КРАС"].Value,Reps["КН РЕЖ1 КРАС"].Value,Reps["КН РЕЖ2 КРАС"].Value)
x=cmdmode_in(4,18,Reps["КОМ РЕЖ3"].Value)


//
// блокировка команд ТУ в режимах АРМ и ПУДП (1-команды разрешены)
// режим грс 0-по месту, 1-пу, 2-арм
//
// упр кранами
//
SET(Reps["ТУ ОБ СПУ"].sys_num, (convertToInteger)
(Reps["РЕЖИМ ГРС КРАС"].Value) == convertToInteger(1))
SET(Reps["ТУ ОБ АРМ"].sys_num, (convertToInteger)
(Reps["РЕЖИМ ГРС КРАС"].Value) == convertToInteger(2))


// режим неуправляемый, при изменениях сохраняем на диск
//
set Reps["АВТОЗАПУСКР"].sys_num,Reps["РЕЖИМ ГРС КРАС"].Value


// если режим пу дп и нет связи с дп - перевести режим в арм
// 300сек- в modbus_s "время обрыва связи"
//
if (convertToInteger(Reps["РЕЖИМ ГРС КРАС"].Value) == convertToInteger(1)) && (convertToInteger(DOST(Reps["СВЯЗ ЛПУ КРАС"].Value)) == convertToInteger(1)) && (convertToInteger(Reps["СВЯЗ ЛПУ КРАС"].Value) == convertToInteger(0)) {
  dout[4]=2
}


// если нет связи с бус в течение 120с - выдать сигнализацию
//
dout[13]=valTrack((convertToInteger(DOST(Reps["СВЯЗЬ С БУС1"].Value)) == convertToInteger(1)) && !Reps["СВЯЗЬ С БУС1"].Value,120,21)
dout[14]=valTrack((convertToInteger(DOST(Reps["СВЯЗЬ С БУС2"].Value)) == convertToInteger(1)) && !Reps["СВЯЗЬ С БУС2"].Value,120,22)

//-------------- засылка расхода газа в одоризатор -------------
//
if (convertToInteger(Reps["КРАН БАЙП КРАС"].Value) == convertToInteger(2)) {
  x=setq_periodic(16,Reps["1SF QМГН КРАС"].Value,Reps["Q ЗАМ1 КРАС"].Value,Reps["РЕЖ ОДОР1 КРАС"].Value,Reps["QАВТ БОМ КРАС"].sys_num,30)
}

}