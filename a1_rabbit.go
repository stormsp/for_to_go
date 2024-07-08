package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
	"math"
)

var aout [100]int
var dout [100]int
var x bool

var InputMap SafeMap
var OutputMap SafeMap

var CONNECTRABBITMIB = "amqp://admin:admin@192.168.1.102:5672/"
var NameAlg = "ButtonALG"

var ConnectToRabit bool
var ConnRabbitMQPublish *amqp.Connection
var ConnRabbitMQConsume *amqp.Connection

type BtnConditionStruct struct {
	Mu                    sync.Mutex
	BtnPressedAndAccident bool      // Флаг того, что у нас кнопка нажата долго и есть авария
	BtnIsRealise          bool      // Флаг того, что кнопка отпущена
	BtnLastPress          time.Time // Когда последний раз нажимали
}

type SafeMap struct {
	Mu   sync.Mutex
	Reps map[string]*Rep
}

type Rep struct {
	MEK_Address int
	Raper       string
	Value       float32
	TypeParam   string
	OldValue    float32
	Reliability bool
	TimeOld     time.Time
	Time        time.Time
}

type OutToRabbitMQ struct {
	MEK_Address int
	Raper       string
	Value       float32
	TypeParam   string
	Reliability bool
	Time        time.Time
}

func main() {
	var err error
	// Устанавливаем соединение для публикации сообщений
	ConnRabbitMQPublish, err = amqp.Dial(CONNECTRABBITMIB)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ for publishing: %v", err)
	}
	defer ConnRabbitMQPublish.Close()

	// Устанавливаем соединение для потребления сообщений
	ConnRabbitMQConsume, err = amqp.Dial(CONNECTRABBITMIB)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ for consuming: %v", err)
	}
	defer ConnRabbitMQConsume.Close()

	// Инициализируем общую структуру с исходными данными
	//safeMap := initializeSafeMap(safeMap)
	InputMap.Reps = make(map[string]*Rep)
	//var InputMap SafeMap
	//fmt.Println(safeMap)
	// Запускаем горутину для потребления сообщений
	fmt.Println(InputMap)
	go ConsumeFromRabbitMq(&InputMap)
	fmt.Println("consume")
	fmt.Println(InputMap)
	//
	// Запускаем горутину для отправки сообщений
	go SendToRabbitMQ(&InputMap)

	// Отправляем тестовое сообщение для проверки потребителя
	//publishTestMessage()



	// Для того чтобы main не завершилась и программа продолжала работать
	fmt.Println("Press [enter] to exit...")
	fmt.Scanln()
}


//___________________


//
// valTrackGt и valTrackLt возращают 0 если отслеживаемый
// параметр не достоверен, или если не нарушена
// граница, или если со времени нарушения
// не прошло timeout секунд. В противном случае
// функции возвращают 1.
//

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
func setwex(sys *float32, value float32, timeout any) bool {
	//fmt.Println(*sys)
	if convertToInteger(SET_WAIT(sys,value,timeout)) != convertToInteger(0) {
		//fmt.Println("voshel")
		//fmt.Println(*sys)
		//time.Sleep((18) * time.Second)
		return(SET_WAIT(sys,value,timeout))
	}
	return(false)
}

//
// impuls
//
func impuls(sys *float32, t any) any {
	x=SET_WAIT(sys,1,t)
	//time.Sleep((2*18) * time.Second)
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



func setwex_DOST(sys *float32, value float32, timeout any) any {
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
func checkFire(Reps *SafeMap) bool {
	//x = 0
	x=x || Reps.Reps["КН РЕЖ1 КРАС"].Value > 0//<Пожар в операторной>.
	//x=x || Reps.Reps["ПОЖАР ОПЕ КРАС"].Value > 0//<Пожар в операторной>.
	//x=x || Reps.Reps["ПОЖАР ПЕР РАС"].Value > 0 //<Пожар в блоке переключения>.
	x=x || Reps.Reps["КН РЕЖ2 КРАС"].Value > 0 //<Пожар в блоке переключения>.
	//x=x || Reps["ПОЖАР ОДО КРАС"].Value //<Пожар в блоке одоризации>.
	return(x)
}


func checkPrecond(Reps *SafeMap) any {
	var x bool
	//x = 0
	if convertToInteger(Reps.Reps["КН РЕЖ1 КРАС"].Value) != convertToInteger(0) {
		//if convertToInteger(Reps.Reps["РЕЖИМ ГРС КРАС"].Value) != convertToInteger(0) {
		x=x||Reps.Reps["КН РЕЖ1 КРАС"].Value > 0 //1 команда - без условий
		//x=x||Reps.Reps["КОМ АО КРАС"].Value > 0 //1 команда - без условий
		if (convertToInteger(valTrack(Reps.Reps["РЕЖИМ ГРС КРАС"].Value, 4, 8)) == convertToInteger(1)) {    // кнопка - только при аварийной ситуации {
			//if (convertToInteger(valTrack(Reps.Reps["КН АВОСТ КРАС"].Value, 4, 8)) == convertToInteger(1)) {    // кнопка - только при аварийной ситуации {
			x=x||checkFire(Reps)        //2 Пожар
			x=x||Reps.Reps["КОМ РЕЖ1 КРАС"].Value  > 0 //3 Аварийно-высокое давление
			//x=x||Reps.Reps["РВЫХ123АВ КРАС"].Value  > 0 //3 Аварийно-высокое давление
		}
	}
	return(x)
}
//--------------------------------------------------------------------------------


//_________________________________________________________________________________

//func initializeSafeMap(safeMap SafeMap) *SafeMap {
//	safeMap.Rep = make(map[string]Rep)
//	return safeMap
//}




// ПС 2 из 3 больше, если достоверен только 1 датчик, не вырабатывать
// mux - множитель типа 110%
// тратит 3 переменные
//


// val,i - значение и индекс измерения
// DOSTval,di - значение и индекс достоверного измерения
// если val достоверно, возвращает индекс ближайшего из двух к среднему
// иначе водвращает di
//


// расчет индекса датчика давления для регулятора на двух эр-04
// 0 - эр04-12, 1 - эр04-21, 2 - эр04-22, подача в том же порядке
//



func er04_neDOST(p any) bool {
	return(!DOST(p) || (convertToInteger(p) < convertToInteger(0.1)) || (convertToInteger(p) > convertToInteger(15.9)))
}

// ********** Функция вычисления qd, qy
// t      - время устройства на этом шаге
// vs_sys - vsum, параметр непрерывный расход (исходный для всех расчетов)
// qf_sys - fix, параметр, где фиксируется непрер накопленный расход
//          при смене контр часа (уст извне, чтобы хранить)
// qy_sys - sys параметра qy (уст извне, чтобы хранить)
// qd_ind - индекс расчетной переменной qd
// qmax   - максимальный возможный расход за сутки, больше - недост
// chour  - контрактный час
// aout[vi+0] - расход с начала суток
// aout[vi+1] - для слежения за изменением времени
//

//-----------------------------------------------------
//Эта функция не используется в самой программе
//func oninit(t any) {
//	INITOUTS(150, 0, 50)
//	// ждем первого опроса модулей
//	time.Sleep((3*18) * time.Second)
//	// автозапуск
//	//
//	SET(Reps.Reps["АЛГ АВОСТ КРАС"].Value, Reps.Reps["АВТОЗАПУСК1"].Value)
//	SET(Reps.Reps["АЛГ БАЙП КРАС"].Value, Reps.Reps["АВТОЗАПУСК2"].Value)
//	//SET(Reps["АЛГ СЛИВ КРАС"].sys_num, Reps["АВТОЗАПУСК3"].Value)
//	SET(Reps.Reps["АЛГ ОТСПГ КРАС"].Value, Reps.Reps["АВТОЗАПУСК2"].Value)
//}

func mainOutput(Reps *SafeMap) {
	SET(Reps.Reps["АВТОЗАПУСК1"].Value,Reps.Reps["АЛГ АВОСТ КРАС"].Value)
	SET(Reps.Reps["АВТОЗАПУСК2"].Value,Reps.Reps["АЛГ БАЙП КРАС"].Value)
	//set Reps["АВТОЗАПУСК3"].sys_num,Reps["АЛГ СЛИВ КРАС"].Value
	SET(Reps.Reps["АВТОЗАПУСК4"].Value,Reps.Reps["АЛГ ОТСПГ КРАС"].Value)
	//--- давление на входе ГРС
	if (convertToInteger(Reps.Reps["РВХ КРАС"].Value) < convertToInteger(Reps.Reps["РВХ МИН КРАС"].Value)) { dout[1]= 1 } else { dout[1]= 0 } // ПС Pвх низ
	if (convertToInteger(Reps.Reps["РВХ КРАС"].Value) > convertToInteger(Reps.Reps["РВХ МАКС КРАС"].Value)) { dout[2] = 1 } else { dout[2] = 0 } // ПС Pвх выс
	if (convertToInteger(Reps.Reps["РВЫХ Д1 КРАС"].Value) < convertToInteger(0.92 * Reps.Reps["РВЫХ ЗАД КРАС"].Value)) { dout[3] = 1 } else { dout[3] = 0 } // ПС Рвых1 пониж по 1 каналу
	if (convertToInteger(Reps.Reps["РВЫХ Д1 КРАС"].Value) > convertToInteger(1.08 * Reps.Reps["РВЫХ ЗАД КРАС"].Value)) { dout[4] = 1 } else { dout[4] = 0 } // ПС Рвых1 повыш по 1 каналу
	if (convertToInteger(Reps.Reps["РВЫХ Д2 КРАС"].Value) < convertToInteger(0.92 * Reps.Reps["РВЫХ ЗАД КРАС"].Value)) { dout[5] = 1 } else { dout[5] = 0 } // ПС Рвых1 низ по 2 каналу
	if (convertToInteger(Reps.Reps["РВЫХ Д2 КРАС"].Value) > convertToInteger(1.08 * Reps.Reps["РВЫХ ЗАД КРАС"].Value)) { dout[6] = 1 } else { dout[6] = 0 } // ПС Рвых1 выс по 2 каналу
	if (convertToInteger(Reps.Reps["РВЫХ Д3 КРАС"].Value) < convertToInteger(0.92 * Reps.Reps["РВЫХ ЗАД КРАС"].Value)) { dout[7] = 1 } else { dout[7] = 0 } // ПС Рвых1 низ по 3 каналу
	if (convertToInteger(Reps.Reps["РВЫХ Д3 КРАС"].Value) > convertToInteger(1.08 * Reps.Reps["РВЫХ ЗАД КРАС"].Value)) { dout[8] = 1 } else { dout[8] = 0 } // ПС Рвых1 выс по 3 каналу
	if (ps2is3Lt(Reps.Reps["РВЫХ Д1 КРАС"].Value, Reps.Reps["РВЫХ Д2 КРАС"].Value, Reps.Reps["РВЫХ Д3 КРАС"].Value, 90, Reps.Reps["РВЫХ ЗАД КРАС"].Value, 10, 150)) { dout[9] = 1 } else { dout[9] = 0 } // давление газа на выходе низкое
	if (ps2is3Gt(Reps.Reps["РВЫХ Д1 КРАС"].Value, Reps.Reps["РВЫХ Д2 КРАС"].Value, Reps.Reps["РВЫХ Д3 КРАС"].Value, 111, Reps.Reps["РВЫХ ЗАД КРАС"].Value, 10, 153)) { dout[10] = 1 } else { dout[10] = 0 } // давление газа на выходе высокое
	if (ps2is3Lt(Reps.Reps["РВЫХ Д1 КРАС"].Value, Reps.Reps["РВЫХ Д2 КРАС"].Value, Reps.Reps["РВЫХ Д3 КРАС"].Value, 87, Reps.Reps["РВЫХ ЗАД КРАС"].Value, 40, 156)) { dout[11] = 1 } else { dout[11] = 0 } // давление газа на выходе предельно-низкое
	if (ps2is3Gt(Reps.Reps["РВЫХ Д1 КРАС"].Value, Reps.Reps["РВЫХ Д2 КРАС"].Value, Reps.Reps["РВЫХ Д3 КРАС"].Value, 112, Reps.Reps["РВЫХ ЗАД КРАС"].Value, 40, 159)) { dout[12] = 1 } else { dout[12] = 0 } // давление газа на выходе предельно-высокое
	if (ps2is3Gt(Reps.Reps["РВЫХ Д1 КРАС"].Value, Reps.Reps["РВЫХ Д2 КРАС"].Value, Reps.Reps["РВЫХ Д3 КРАС"].Value, 110, Reps.Reps["РВЫХ ЗАД КРАС"].Value, 10, 162)) { dout[13] = 1 } else { dout[13] = 0 } // давление газа на выходе опасно-высокое
	if (ps2is3Gt(Reps.Reps["РВЫХ Д1 КРАС"].Value, Reps.Reps["РВЫХ Д2 КРАС"].Value, Reps.Reps["РВЫХ Д3 КРАС"].Value, 115, Reps.Reps["РВЫХ ЗАД КРАС"].Value, 10, 165)) { dout[14] = 1 } else { dout[14] = 0 } // давление газа на выходе аварийно-высокое
	if (convertToInteger(Reps.Reps["ТВЫХ КРАС"].Value) < convertToInteger(-10)) { dout[15] = 1 } else { dout[15] = 0 } // Твых1 низ
	if (convertToInteger(Reps.Reps["ТВЫХ КРАС"].Value) < convertToInteger(0)) { dout[16] = 1 } else { dout[16] = 0 } // Твых1 пониж
	if (convertToInteger(Reps.Reps["ПГ ТВЫХ КРАС"].Value) > convertToInteger(60)) { dout[17] = 1 } else { dout[17] = 0 } // Твых1 повыш

	//--- температура в помещениях
	if (valTrackGt(Reps.Reps["ТВЗД ОПЕР КРАС"].Value, 70, 10, 168)) { dout[18] = 1 } else { dout[18] = 0 } // Твоз опер высокая
	if (valTrackLt(Reps.Reps["ТВЗД ОПЕР КРАС"].Value, 5, 10, 169)) { dout[19] = 1 } else { dout[19] = 0 } // Твоз опер низкая
	//if (valTrackGt(Reps.Reps["ТВЗД ТОПЧ КРАС"].Value, 70, 10, 170)) { dout[20] = 1 } else { dout[20] = 0 } // Твоз топоч высокая
	//if (valTrackLt(Reps.Reps["ТВЗД ТОПЧ КРАС"].Value, 5, 10, 171)) { dout[21] = 1 } else { dout[21] = 0 } // Твоз топоч низкая
	if (valTrackGt(Reps.Reps["ТВЗД ПЕР КРАС"].Value, 70, 10, 172)) { dout[22] = 1 } else { dout[22] = 0 } // Твоз перкл высокая
	if (valTrackGt(Reps.Reps["ТВЗД РЕД КРАС"].Value, 70, 10, 173)) { dout[23] = 1 } else { dout[23] = 0 } // Твоз редуц высокая
	//if (valTrackGt(Reps.Reps["ТВЗД РАСХ КРАС"].Value, 70, 10, 174)) { dout[24] = 1 } else { dout[24] = 0 } // Твоз расходм высокая
	//if (valTrackGt(Reps.Reps["ТВЗД ОДО КРАС"].Value, 70, 10, 175)) { dout[25] = 1 } else { dout[25] = 0 } // Твоз одор высокая
	if (valTrack((convertToInteger(Reps.Reps["ЯХПЖ ОПЕР КРАС"].Value) == convertToInteger(2)) && convertToBool(Reps.Reps["ТВ ОПЕР В КРАС"].Value), 10, 176)) { dout[26] = 1 } else { dout[26] = 0 } // пожар в операторной
	//if (valTrack((convertToInteger(Reps.Reps["ЯХПЖ ТОПЧ КРАС"].Value) == convertToInteger(2)) && Reps.Reps["ТВ ТОП В КРАС"].Value, 10, 177)) { dout[27] = 1 } else { dout[27] = 0 } // пожар в топочной
	if (valTrack((convertToInteger(Reps.Reps["ЯХПЖ ПЕР КРАС"].Value) == convertToInteger(2)) && convertToBool(Reps.Reps["ТВ ПЕР В КРАС"].Value), 10, 178)) { dout[28] = 1 } else { dout[28] = 0 } // пожар в бл.переключ
	if (valTrack((convertToInteger(Reps.Reps["ЯХПЖ РЕД КРАС"].Value) == convertToInteger(2)) && convertToBool(Reps.Reps["ТВ РЕД В КРАС"].Value), 10, 179)) { dout[29] = 1 } else { dout[29] = 0 } // пожар в бл.редуц
	//if (valTrack((convertToInteger(Reps.Reps["ЯХПЖ РАСХ КРАС"].Value) == convertToInteger(2)) && Reps.Reps["ТВ РАСХ В КРАС"].Value, 10, 180)) { dout[30] = 1 } else { dout[30] = 0 } // пожар в расходомерной
	//if (valTrack((convertToInteger(Reps.Reps["ЯХПЖ ОДОР КРАС"].Value) == convertToInteger(2)) && Reps.Reps["ТВ ОДОР В КРАС"].Value, 10, 181)) { dout[31] = 1 } else { dout[31] = 0 } // пожар в бл.одоризации
	if (convertToInteger(Reps.Reps["DP ФИЛЬТ1 КРАС"].Value) > convertToInteger(Reps.Reps["DР ФЛ ПВ КРАС"].Value)) { dout[32] = 1 } else { dout[32] = 0 } // перепад на фильтре1 повыш
	if (convertToInteger(Reps.Reps["DP ФИЛЬТ1 КРАС"].Value) > convertToInteger(Reps.Reps["DР ФЛ В КРАС"].Value)) { dout[33] = 1 } else { dout[33] = 0 } // перепад на фильтре1 выс
	if (convertToInteger(Reps.Reps["DP ФИЛЬТ2 КРАС"].Value) > convertToInteger(Reps.Reps["DР ФЛ ПВ КРАС"].Value)) { dout[34] = 1 } else { dout[34] = 0 } // перепад на фильтре2 повыш
	if (convertToInteger(Reps.Reps["DP ФИЛЬТ2 КРАС"].Value) > convertToInteger(Reps.Reps["DР ФЛ В КРАС"].Value)) { dout[35] = 1 } else { dout[35] = 0 } // перепад на фильтре2 выс
	if (convertToInteger(Reps.Reps["РСВ ЛИН КРАС"].Value) > convertToInteger(Reps.Reps["РСВЛ МАКС КРАС"].Value)) { dout[36] = 1 } else { dout[36] = 0 } // открытие ппк
	if (valTrackGt(Reps.Reps["1SF dР КРАС"].Value, Reps.Reps["DРУГ1МАКС КРАС"].Value, 20, 182)) { dout[37] = 1 } else { dout[37] = 0 } // перепад выс
	if (valTrackLt(Reps.Reps["1SF dР КРАС"].Value, Reps.Reps["DРУУГ1МИН КРАС"].Value, 20, 183)) { dout[38] = 1 } else { dout[38] = 0 } // перепад низ
	if (valTrackGt(Reps.Reps["1SF Р КРАС"].Value, Reps.Reps["РУУГ1МАКС КРАС"].Value, 20, 184)) { dout[39] = 1 } else { dout[39] = 0 } // давление выс
	if (valTrackLt(Reps.Reps["1SF Р КРАС"].Value, Reps.Reps["РУУГ1МИН КРАС"].Value, 20, 185)) { dout[40] = 1 } else { dout[40] = 0 } // давление низ
	if (convertToInteger(Reps.Reps["РРЕГ1 КРАС"].Value) > convertToInteger(Reps.Reps["РРЕГ1 ВТГ КРАС"].Value)) { dout[41] = 1 } else { dout[41] = 0 }
	if (convertToInteger(Reps.Reps["РРЕГ1 КРАС"].Value) > convertToInteger(Reps.Reps["РРЕГ1 ВАГ КРАС"].Value)) { dout[41] += 1 } // давление на регуляторе (на командном газе)
	if (convertToInteger(Reps.Reps["РРЕГ2 КРАС"].Value) > convertToInteger(Reps.Reps["РРЕГ2 ВТГ КРАС"].Value)) { dout[42] = 1 } else { dout[42] = 0 }
	if (convertToInteger(Reps.Reps["РРЕГ2 КРАС"].Value) > convertToInteger(Reps.Reps["РРЕГ2 ВАГ КРАС"].Value)) { dout[42] += 1 } // давление на регуляторе (на командном газе)
	if (!DOST(Reps.Reps["ПОЗ ЗАДВ КРАС"].Value)) { dout[44] = 1 } else { dout[44] = 0 } // исправность каналов эр-04 4 шт
	if (er04_neDOST(Reps.Reps["РВЫХ Д1 КРАС"].Value)) { dout[45] = 1 } else { dout[45] = 0 }
	if (er04_neDOST(Reps.Reps["РВЫХ Д2 КРАС"].Value)) { dout[46] = 1 } else { dout[46] = 0 }
	if (er04_neDOST(Reps.Reps["РВЫХ Д3 КРАС"].Value)) { dout[47] = 1 } else { dout[47] = 0 }


	//---- состояние грс
	//#include "KRASNOUS\FORM\sost_grs.evl"
	//dout[48]=sost_grs(0)

	//----------- переcчет в кг --------
	aout[49]=int(Reps.Reps["РВХ КРАС"].Value*10.19716)
	aout[50]=int(Reps.Reps["РВЫХ Д1 КРАС"].Value*10.19716)
	aout[51]=int(Reps.Reps["РВЫХ Д2 КРАС"].Value*10.19716)
	aout[52]=int(Reps.Reps["РВЫХ Д3 КРАС"].Value*10.19716)
	aout[53]=int(Reps.Reps["РВЫХ123 КРАС"].Value*10.19716)

	//------ Расчет обьема газа при работе на байпасе -----
	Vbp(Reps.Reps["КРАН БАЙП КРАС"].Value,Reps.Reps["1SF ПРСУТ КРАС"].Value,54) // 5переменных


	// большой расход на байпасе
	//
	if (valTrackGt(Reps.Reps["ПОЗ ЗАДВ КРАС"].Value,97,120,192)) {dout[66]=1} else {dout[66]=0}

	// расчет индекса канала давления для регулятора на двух эр-04
	//x = 0
	x=x || (convertToInteger(Reps.Reps["2ЭР04 КРАС"].Value) == convertToInteger(0))
	x=x || (convertToInteger(DOST(Reps.Reps["РВЫХ Д2 КРАС"].Value) && DOST(Reps.Reps["РВЫХ Д3 КРАС"].Value)) == convertToInteger(0))

	if x {
		dout[67]=0
	} else {
		dout[67]=int(regp3i(Reps.Reps["РВЫХ Д1 КРАС"].Value,Reps.Reps["РВЫХ Д2 КРАС"].Value,Reps.Reps["РВЫХ Д3 КРАС"].Value,float32(dout[59])))
	}

	// если разница между сигналом управления и фактич положением задвижки велика
	// значит Роторк застрял
	if (valTrackGt(float32(math.Abs(float64(Reps.Reps["ПОЗ ЗАДВ КРАС"].Value-Reps.Reps["УПР ЗАДВ КРАС"].Value))),5,90,191)) {dout[68]=1} else {dout[68]=0}


	//--------------- sevc ------------------------------------

	upd_qyqd(&Reps.Reps["ВРЕМЯSEVC КРАС"].Value,&Reps.Reps["VСУМ SEVC КРАС"].Value,&Reps.Reps["VНАС SEVC КРАС"].Value,&Reps.Reps["QY SEVC КРАС"].Value,61,5000,12)

	//--------------- БОМ ОВЕН---------------------------------
	bomOven(121,Reps.Reps["БОМСТАТ1 КРАС"].Value,Reps.Reps["БОМСТАТ2 КРАС"].Value)
	//---- ИБП
	if (valTrackGt(Reps.Reps["I НАГ ИБП КРАС"].Value,0.9*Reps.Reps["IМАКС ИБП КРАС"].Value,2,186)) {dout[72]=1} else {dout[72]=0} // ток нагрузки ибп высокий (номинально smart 30А)
	if (convertToInteger(Reps.Reps["Т ИНВ ИБП КРАС"].Value) > convertToInteger(70)) {
		dout[73] = 1; // перегрев ибп
	} else {
		dout[73] = 0;
	}

	if (convertToInteger(Reps.Reps["UВХ ИБП КРАС"].Value) < convertToInteger(185)) {
		dout[74] = 1; // u вх ибп низкое
	} else {
		dout[74] = 0;
	}

	if (convertToInteger(Reps.Reps["UВХ ИБП КРАС"].Value) > convertToInteger(245)) {
		dout[75] = 1; // u вх ибп высокое
	} else {
		dout[75] = 0;
	}//---------------- Яхонт 4и -------------------------
	yahont4(76,Reps.Reps["ЯХ ШЛ1"].Value)
	yahont4(77,Reps.Reps["ЯХ ШЛ2"].Value)
	yahont4(78,Reps.Reps["ЯХ ШЛ3"].Value)
	yahont4(79,Reps.Reps["ЯХ ШЛ4"].Value)
	yahont41(80,Reps.Reps["ЯХ КОРПУС"].Value)
	yahont42(81,Reps.Reps["ЯХ ОСНОВН ИП"].Value)
	yahont42(82,Reps.Reps["ЯХ РЕЗЕРВ ИП"].Value)

	// --- уставки УКЗ
	aout[83]=int(121.9+Reps.Reps["СКЗ ИМПУЛ КРАС"].Value/3200)
	if (valTrackGt(Reps.Reps["СКЗ E КРАС"].Value, Reps.Reps["EСКЗ МАКС КРАС"].Value, 10, 187)) { dout[84] = 1; } else { dout[84] = 0; } // пот скз высокий
	if (valTrackLt(Reps.Reps["СКЗ E КРАС"].Value, Reps.Reps["EСКЗ МИН КРАС"].Value, 10, 188)) { dout[85] = 1; } else { dout[85] = 0; } // пот скз низкий
	if (valTrackGt(Reps.Reps["СКЗ I КРАС"].Value, Reps.Reps["IСКЗ МАКС КРАС"].Value, 10, 189)) { dout[86] = 1; } else { dout[86] = 0; } // ток скз высокий
	if (valTrackGt(Reps.Reps["СКЗ U КРАС"].Value, Reps.Reps["UСКЗ МАКС КРАС"].Value, 10, 190)) { dout[87] = 1; } else { dout[87] = 0; } // u скз высокое

	//-------------- СГОЭС --------------------------
	dout[88]=sgoes_avar(int(Reps.Reps["1СГОЭС КОД"].Value))
	if (convertToInteger(Reps.Reps["СН4 ПЕР КРАС"].Value) > convertToInteger(10)) {dout[89]= 1} else {dout[89]= 0}
	if (convertToInteger(Reps.Reps["СН4 ПЕР КРАС"].Value) > convertToInteger(20)) {dout[90]= 1} else {dout[90]= 0}
	dout[91]=sgoes_avar(int(Reps.Reps["2СГОЭС КОД"].Value))
	if (convertToInteger(Reps.Reps["СН4 РЕД КРАС"].Value) > convertToInteger(10)) {dout[92]=1} else {dout[92]= 0}
	if (convertToInteger(Reps.Reps["СН4 РЕД КРАС"].Value) > convertToInteger(20)) {dout[93]=1} else {dout[93]= 0}
	// -------------- БУПГ24-3 -----------------------
	bupg243(100,Reps.Reps["ПГ СОСТ"].Value)
	bupg243_klap(114,Reps.Reps["ПГ СОС КЛ"].Value)
}


func ConsumeFromRabbitMq(Reps *SafeMap) {
	//Conn := ConnRabbitMQConsume
	ch, err := ConnRabbitMQConsume.Channel()
	if err != nil {
		fmt.Println("Ошибка открытия канала RabbitMQ ", err)
	}

	defer ch.Close()
	args := amqp.Table{
		"x-max-length": 1,
		"x-overflow":   "reject-publish",
	}
	q, err := ch.QueueDeclare(
		NameAlg, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		args,    // arguments
	)
	if err != nil {
		fmt.Println("Consumer Ошибка декларирования очереди RabbitMQ ", NameAlg+"Out", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		fmt.Println("Consumer Ошибка Qos RabbitMQ ", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		args,   // args
	)
	if err != nil {
		fmt.Println("Consumer Ошибка создания Consumer ", err)
	}

	var forever chan struct{}

	if err == nil {
		MessageHandler(msgs, Reps)
	}
	fmt.Println(" [*] Waiting for messages.")
	<-forever
}

func MessageHandler(msgs <-chan amqp.Delivery, Reps *SafeMap) {
	if Reps == nil {
		fmt.Println("Reps is nil in MessageHandler")
		return
	}
	for d := range msgs {
		fmt.Println("Message received:", string(d.Body)) // Logging received messages
		var data []Rep
		err := json.Unmarshal(d.Body, &data)
		if err != nil {
			fmt.Println("Ошибка разбора JSON:", err)
			continue
		}
		// ************ ЗАПИСЬ В ОБЩУЮ СТРУКТУРУ **********
		Reps.Mu.Lock()
		for _, inputVal := range data {
			ConnectToRabit = true
			repVal, exist := Reps.Reps[inputVal.Raper]
			if exist {
				fmt.Printf("Updating rep %s: %v -> %v\n", inputVal.Raper, repVal.Value, inputVal.Value) // Logging updates
				repVal.Value = inputVal.Value
				repVal.Time = inputVal.Time
			} else {
				fmt.Println("Adding new rep to map:", inputVal.Raper) // Logging new entries
				Reps.Reps[inputVal.Raper] = &Rep{
					Value:       inputVal.Value,
					Time:        inputVal.Time,
					Raper:       inputVal.Raper,
					MEK_Address: inputVal.MEK_Address,
					TypeParam:   inputVal.TypeParam,
				}
			}
		}
		Reps.Mu.Unlock()

		//fmt.Println(Reps.Reps["КРАН ОХР КРАС"].Value)
		fmt.Println("выполняю mainOutput")
		// Выполняем основную логику обработки
		fmt.Println(Reps)
		mainOutput(Reps)
		fmt.Println("готово")
		//fmt.Println(Reps.Reps["КРАН ОХР КРАС"].Value)

		fmt.Println("Updated Reps:", Reps)
		// Отправляем измененные реперы обратно в RabbitMQ
		//SendToRabbitMQ(Reps)
		d.Ack(false)
	}
}


// SendToRabbitMQ отправка Структуры в очередь по названию (для мэк)
func SendToRabbitMQ(OutputMap *SafeMap) {

	for {
		OutputMap.Mu.Lock()
		output := OutputMap.Reps
		var outToRabbit = make([]OutToRabbitMQ, 0)
		for key, _ := range output {
			value := output[key]
			if value.TimeOld != value.Time {
				outToRabbit = append(outToRabbit, OutToRabbitMQ{value.MEK_Address, value.Raper, value.Value, value.TypeParam, value.Reliability, value.Time})
				outVal, exist := OutputMap.Reps[key]
				if exist {
					outVal.TimeOld = outVal.Time
					OutputMap.Reps[key] = outVal
				}
			}
		}
		OutputMap.Mu.Unlock()
		if len(outToRabbit) > 0 {
			body, err := json.Marshal(outToRabbit)
			if err != nil {
				fmt.Println("Ошибка При формировании JSON ", err)
			}
			//fmt.Println("______________________________________________________________________Outtorabbit_____________________________")
			//fmt.Println(outToRabbit)
			ch, err := ConnRabbitMQPublish.Channel()
			if err != nil {
				fmt.Println("Ошибка открытия канала RabbitMQ ", err)
			}
			args := amqp.Table{
				"x-max-length": 1,
				"x-overflow":   "reject-publish",
			}
			q, err := ch.QueueDeclare(
				NameAlg+"Out", // name
				false,         // durable
				false,         // delete when unused
				false,         // exclusive
				false,         // no-wait
				args,          // arguments
			)
			if err != nil {
				fmt.Println("Failed to declare a queue ", err)

			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err = ch.PublishWithContext(ctx,
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
			if err != nil {
				fmt.Println("Ошибка отправки в очередь", err)
			}
			ch.Close()
			cancel()
			//fmt.Println(" [x] Отправил в очередь ", outToRabbit)
		}
		time.Sleep(100 * time.Millisecond)
	}
}


// Function to publish a test message
func publishTestMessage() {
	ch, err := ConnRabbitMQPublish.Channel()
	if err != nil {
		fmt.Println("Ошибка открытия канала для отправки тестового сообщения RabbitMQ ", err)
		return
	}
	defer ch.Close()

	testData := map[string]Rep{
		"КН АВОСТ КРАС": {MEK_Address: 1, Raper: "КН АВОСТ КРАС", Value: 150, TypeParam: "param1", Reliability: true, Time: time.Now()},
		// Add more test data as needed
	}

	body, err := json.Marshal(testData)
	if err != nil {
		fmt.Println("Ошибка При формировании тестового JSON ", err)
		return
	}

	args := amqp.Table{
		"x-max-length": 1,
		"x-overflow":   "reject-publish",
	}
	q, err := ch.QueueDeclare(
		NameAlg, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		args,    // arguments
	)
	if err != nil {
		fmt.Println("Failed to declare a queue ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		fmt.Println("Ошибка отправки тестового сообщения в очередь", err)
	} else {
		fmt.Println("Тестовое сообщение отправлено успешно")
	}
}
