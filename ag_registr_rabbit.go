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


//ba+0 - запрос снизу
//ba+1 - ответ сверху
//ba+2 - результат. 0- запрет, 1 - ожидание ответа, 2 - разрешение
//id   - для слежения

func dopusk(ba int, id int) {
	if convertToInteger(dout[ba]) != convertToInteger(dout[id]) {
		if convertToInteger(dout[ba]) == convertToInteger(1) { //запрос снизу {
			dout[ba+2]=1 //сигнал о запросе наверх
		}
		if convertToInteger(dout[ba]) == convertToInteger(2) { //завершено снизу {
			dout[ba+2]=0 //
		}
	}
	if (convertToInteger(dout[ba]) == convertToInteger(1))  &&  (convertToInteger(dout[ba+1]) == convertToInteger(1)) {//разрешение получено {
		dout[ba+0]=0 //сброс
		dout[ba+1]=0 //сброс
		dout[ba+2]=2 //разрешен
	}

	if convertToInteger(dout[ba+1]) == convertToInteger(2) {//получен запрет {
		dout[ba+0]=0 //сброс
		dout[ba+1]=0 //сброс
		dout[ba+2]=0 //запрет
	}
}

//таймер возвращает время активной comm в секундах
// id - aout для работы
// comm - команда управления 1-считать, иначе сброс
func timer(id int, comm any) any {
	if convertToInteger(comm) == convertToInteger(1) {
		if convertToInteger(aout[id]) == convertToInteger(0) {
			aout[id]=int(time.Now().Unix())
		}
		return(int(time.Now().Unix())-aout[id])
	}

	aout[id]=0
	return(false)
}


// функция для фиксирования факта появления сигнала
// если cond=1 , то бит установится иначе без изменений

func setbitif (cond bool, dword float32, bitnum float32) any {
	if cond {
	return SETBITS(uint32(dword), 1, uint(bitnum),1)
}
	return dword
}

// ------ функция слежения за алгоритмами
func testTrack(reason *float32, error int, status any, t1 int, id1 int, id2 int) any {
	if front(reason,id2) {
	dout[id1+2]= 1 //выполняется
}

	if convertToInteger(dout[id1+2]) == convertToInteger(1) {//выполняется {
	aout[id1+0]= t1
	aout[id1+1]=int(time.Now().Unix())-t1
		if convertToInteger(status) == convertToInteger(0) { //выполнялось {
		dout[id1+2]=2+error
	}
	}
		return(false)
}


func word2outs(w int, id int) {
	dout[id+0]=w & (1 << 0)
	dout[id+1]=w & (1 << 1)
	dout[id+2]=w & (1 << 2)
	dout[id+3]=w & (1 << 3)
	dout[id+4]=w & (1 << 4)
	dout[id+5]=w & (1 << 5)
	dout[id+6]=w & (1 << 6)
	dout[id+7]=w & (1 << 7)
	dout[id+8]=w & (1 << 8)
	dout[id+9]=w & (1 << 9)
	dout[id+10]=w & (1 << 10)
	dout[id+11]=w & (1 << 11)
	dout[id+12]=w & (1 << 12)
	dout[id+13]=w & (1 << 13)
	dout[id+14]=w & (1 << 14)
	dout[id+15]=w & (1 << 15)
}

//-----------------------------------
func oninit(t any) {
	//codps:=0
	//codz:=0
	dout[3]=0
	dout[6]=0
	dout[9]=0
	dout[12]=0
	dout[15]=0
	aout[70]=0
	dout[90]=0
	dout[91]=0
	dout[92]=0
	dout[95]=0
	dout[96]=0
	dout[97]=0
	dout[98]=0
	time.Sleep((18*4) * time.Second)

	//aout[21]=TRUE(Reps.Reps["НАЧ ОПР1 КРАС"].Value)
	//aout[22]=TRUE(Reps.Reps["ВР ОПР1 КРАС"].Value/10)
	//dout[23]=TRUE(Reps.Reps["РЕЗ ОПР1 КРАС"].Value)
	//aout[24]=TRUE(Reps.Reps["НАЧ ОПР2 КРАС"].Value)
	//aout[25]=TRUE(Reps.Reps["ВР ОПР2 КРАС"].Value/10)
	//dout[26]=TRUE(Reps.Reps["РЕЗ ОПР2 КРАС"].Value)
	//aout[27]=TRUE(Reps.Reps["НАЧ ОПР3 КРАС"].Value)
	//aout[28]=TRUE(Reps.Reps["ВР ОПР3 КРАС"].Value/10)
	//dout[29]=TRUE(Reps.Reps["РЕЗ ОПР3 КРАС"].Value)
	//aout[30]=TRUE(Reps.Reps["НАЧ ОПР4 КРАС"].Value)
	//aout[31]=TRUE(Reps.Reps["ВР ОПР4 КРАС"].Value/10)
	//dout[32]=TRUE(Reps.Reps["РЕЗ ОПР4 КРАС"].Value)
	//aout[33]=TRUE(Reps.Reps["НАЧ ОПР5 КРАС"].Value)
	//aout[34]=TRUE(Reps.Reps["ВР ОПР5 КРАС"].Value/10)
	//dout[35]=TRUE(Reps.Reps["РЕЗ ОПР5 КРАС"].Value)
	//aout[36]=TRUE(Reps.Reps["НАЧ ОПР6 КРАС"].Value)
	//aout[37]=TRUE(Reps.Reps["ВР ОПР6 КРАС"].Value/10)
	//dout[38]=TRUE(Reps.Reps["РЕЗ ОПР6 КРАС"].Value)
	//aout[39]=TRUE(Reps.Reps["НАЧ ОПР7 КРАС"].Value)
	//aout[40]=TRUE(Reps.Reps["ВР ОПР7 КРАС"].Value/10)
	//dout[41]=TRUE(Reps.Reps["РЕЗ ОПР7 КРАС"].Value)
	//aout[42]=TRUE(Reps.Reps["НАЧ ОПР8 КРАС"].Value)
	//aout[43]=TRUE(Reps.Reps["ВР ОПР8 КРАС"].Value/10)
	//dout[44]=TRUE(Reps.Reps["РЕЗ ОПР8 КРАС"].Value)
	//aout[45]=TRUE(Reps.Reps["НАЧ ОПР9 КРАС"].Value)
	//aout[46]=TRUE(Reps.Reps["ВР ОПР9 КРАС"].Value/10)
	//dout[47]=TRUE(Reps.Reps["РЕЗ ОПР9 КРАС"].Value)
	//aout[48]=TRUE(Reps.Reps["НАЧ ОПР10 КРАС"].Value)
	//aout[49]=TRUE(Reps.Reps["ВР ОПР10 КРАС"].Value/10)
	//dout[50]=TRUE(Reps.Reps["РЕЗ ОПР10 КРАС"].Value)
	//aout[51]=TRUE(Reps.Reps["НАЧ ОПР11 КРАС"].Value)
	//aout[52]=TRUE(Reps.Reps["ВР ОПР11 КРАС"].Value/10)
	//dout[53]=TRUE(Reps.Reps["РЕЗ ОПР11 КРАС"].Value)
	//aout[54]=TRUE(Reps.Reps["НАЧ ОПР12 КРАС"].Value)
	//aout[55]=TRUE(Reps.Reps["ВР ОПР12 КРАС"].Value/10)
	//dout[56]=TRUE(Reps.Reps["РЕЗ ОПР12 КРАС"].Value)
	//aout[57]=TRUE(Reps.Reps["НАЧ ОПР13 КРАС"].Value)
	//aout[58]=TRUE(Reps.Reps["ВР ОПР13 КРАС"].Value/10)
	//dout[59]=TRUE(Reps.Reps["РЕЗ ОПР13 КРАС"].Value)
	//aout[60]=TRUE(Reps.Reps["НАЧ ОПР14 КРАС"].Value)
	//aout[61]=TRUE(Reps.Reps["ВР ОПР14 КРАС"].Value/10)
	//dout[62]=TRUE(Reps.Reps["РЕЗ ОПР14 КРАС"].Value)
	aout[21]=1
	aout[22]=1
	dout[23]=1
	aout[24]=1
	aout[25]=1
	dout[26]=1
	aout[27]=1
	aout[28]=1
	dout[29]=1
	aout[30]=1
	aout[31]=1
	dout[32]=1
	aout[33]=1
	aout[34]=1
	dout[35]=1
	aout[36]=1
	aout[37]=1
	dout[38]=1
	aout[39]=1
	aout[40]=1
	dout[41]=1
	aout[42]=1
	aout[43]=1
	dout[44]=1
	aout[45]=1
	aout[46]=1
	dout[47]=1
	aout[48]=1
	aout[49]=1
	dout[50]=1
	aout[51]=1
	aout[52]=1
	dout[53]=1
	aout[54]=1
	aout[55]=1
	dout[56]=1
	aout[57]=1
	aout[58]=1
	dout[59]=1
	aout[60]=1
	aout[61]=1
	dout[62]=1
}

// ------------------- Регламентные работы -----------------------------

func testTrackRR(on *float32, off int, id1 int, id2 int) any {
	if front(on,id2) {
		dout[id1+2]= 1 //выполняется
		aout[id1+0]= int(time.Now().Unix())
	}

	if convertToInteger(dout[id1+2]) == convertToInteger(1) {//выполняется {
		aout[id1+1]=int(time.Now().Unix())-aout[id1+0]
		dout[id1+2]=1+off //выполнено
	}
	return(false)
}





func mainOutput(Reps *SafeMap) {
	dopusk( 1,111)
	dopusk( 4,112)
	dopusk( 7,113)
	dopusk(10,114)
	dopusk(13,115)
	dopusk(16,116)


	// слежение за кранами 16 бит(с запасом)
	w:=uint32(aout[70])
	w=SETBITS(w, 1, 0,valTrack(convertToInteger(Reps.Reps["ОК УРЛ КРАС"].Value) != convertToInteger(0),30,100)) //проблемы с охр краном
	w=SETBITS(w, 1, 1,valTrack(convertToInteger(Reps.Reps["КРБП УРЛ КРАС"].Value) != convertToInteger(0),15,101)) //проблемы с БП краном
	w=SETBITS(w, 1, 2,valTrack(convertToInteger(Reps.Reps["КРВХ УРЛ КРАС"].Value) != convertToInteger(0),15,102)) //проблемы с вх краном
	w=SETBITS(w, 1, 3,valTrack(convertToInteger(Reps.Reps["КРВЫХ УРЛ КРАС"].Value) != convertToInteger(0),15,103)) //проблемы с вых краном
	w=SETBITS(w, 1, 4,valTrack(convertToInteger(Reps.Reps["КРЕД1 УРЛ КРАС"].Value) != convertToInteger(0),15,104)) //проблемы с лред краном
	w=SETBITS(w, 1, 5,valTrack(convertToInteger(Reps.Reps["КРЕД2 УРЛ КРАС"].Value) != convertToInteger(0),15,105)) //проблемы с лред краном
	w=SETBITS(w, 1, 6,valTrack(convertToInteger(Reps.Reps["КСВН УРЛ КРАС"].Value) != convertToInteger(0),15,106)) //проблемы с свечн краном (при опробовании это м.б. норма)
	w=SETBITS(w, 1, 7,valTrack(convertToInteger(Reps.Reps["КСВВ УРЛ КРАС"].Value) != convertToInteger(0),15,107)) //проблемы с свечн краном (при опробовании это м.б. норма)
	w=SETBITS(w, 1, 8,valTrack(convertToInteger(Reps.Reps["ПГБП УРЛ КРАС"].Value) != convertToInteger(0),15,108)) //проблемы с байп ПГ
	w=SETBITS(w, 1, 9,valTrack(convertToInteger(Reps.Reps["ПГВХ УРЛ КРАС"].Value) != convertToInteger(0),15,109)) //проблемы с вх ПГ
	w=SETBITS(w, 1, 10,valTrack(convertToInteger(Reps.Reps["ПГВЫХ УРЛ КРАС"].Value) != convertToInteger(0),15,110)) //проблемы с вых ПГ
	aout[70]=w
	x=word2outs(w,71)
	// ------- для байпаса -------------
	//давление в норме 5%
	pNorm=(convertToInteger(Reps.Reps["РВЫХ123 КРАС"].Value) > convertToInteger(0.95*Reps.Reps["РВЫХ ЗАД КРАС"].Value))
	pNorm=pNorm  &&  (convertToInteger(Reps.Reps["РВЫХ123 КРАС"].Value) < convertToInteger(1.05*Reps.Reps["РВЫХ ЗАД КРАС"].Value))
	dout[91]=!pNorm

	// время работы на контуре ЭР-04(КРБП)
	x = convertToInteger(Reps.Reps["КРАН БАЙП КРАС"].Value) != convertToInteger(2)  &&  Reps.Reps["РЕЖ РЕГ04 КРАС"].Value // контур ЭР-04 ( ибель) в работе
	aout[88]=timer(126,x)
	if convertToInteger(aout[88]) != convertToInteger(0) {
		dout[92]=(convertToInteger(aout[88]) < convertToInteger(60*5)) //время на бп меньше
	}

	// время работы на контуре САУ
	x = convertToInteger(Reps.Reps["КРАН БАЙП КРАС"].Value) != convertToInteger(2)  &&  Reps.Reps["РЕЖ РСАУ КРАС"].Value // контур САУ в работе
	aout[89]=timer(127,x)


	// ---------------------------- проверки байпаса --------------------

	if convertToInteger(Reps.Reps["РАЗР1 ДП КРАС"].Value) == convertToInteger(2) { //подтверждено диспетчером {

		// причины безуспешного проведения
		// - время перестановки кранов велико
		// - погрешность регулирования давления более 5%
		// - время работы на байпасе недостаточно
		dout[90]=convertToInteger(Reps.Reps["КОД О  БП КРАС"].Value) != convertToInteger(0) // здесь пока только краны
		error=dout[90] || dout[91] || dout[92]

		x=testTrack((convertToInteger(Reps.Reps["ПРИЧ БП КРАС"].Value) == convertToInteger(1)),error,Reps.Reps["ХОД БАЙП КРАС"].Value,Reps.Reps["ДАТА БП КРАС"].Value,21,117)// по низкому
		x=testTrack((convertToInteger(Reps.Reps["ПРИЧ БП КРАС"].Value) == convertToInteger(2)),error,Reps.Reps["ХОД БАЙП КРАС"].Value,Reps.Reps["ДАТА БП КРАС"].Value,24,118)// по высокому
		x=testTrack((convertToInteger(Reps.Reps["ПРИЧ БП КРАС"].Value) == convertToInteger(3)),error,Reps.Reps["ХОД БАЙП КРАС"].Value,Reps.Reps["ДАТА БП КРАС"].Value,27,119)// по пожару
		x=testTrack((convertToInteger(Reps.Reps["ПРИЧ БП КРАС"].Value) == convertToInteger(4)),error,Reps.Reps["ХОД БАЙП КРАС"].Value,Reps.Reps["ДАТА БП КРАС"].Value,30,120)// по кнопке
	}

	// --------------------- Проверка АО --------------------------------------------
	if convertToInteger(Reps.Reps["РАЗР4 ДП КРАС"].Value) == convertToInteger(2) { //подтверждено диспетчером {

		// причины безуспешного проведения
		// не поданы команды на соответствующие краны
		aout[99]=setbitif (convertToInteger(Reps.Reps["ОК УРЛ КРАС"].Value) == convertToInteger(2),aout[99],0)
			aout[99]=setbitif (convertToInteger(Reps.Reps["КРВХ УРЛ КРАС"].Value) == convertToInteger(2),aout[99],1)
				aout[99]=setbitif (convertToInteger(Reps.Reps["КРВЫХ УРЛ КРАС"].Value) == convertToInteger(2),aout[99],2)
					if front(convertToInteger(Reps.Reps["РЕЖИМ ГРС КРАС"].Value) != convertToInteger(0),132) {
						aout[99]=0
					}
					dout[97]=convertToInteger(aout[99]) != convertToInteger(7)
					dout[98]=convertToInteger(Reps.Reps["РЕЖИМ ГРС КРАС"].Value) != convertToInteger(0)   // режим ГРС -  !0
					error = dout[97] || dout[98]

					x=testTrack((convertToInteger(Reps.Reps["ПРИЧ АО КРАС"].Value) == convertToInteger(2)),error,Reps.Reps["ХОД АВОСТ КРАС"].Value,Reps.Reps["ДАТА АО КРАС"].Value,36,121)// по пожару
					x=testTrack((convertToInteger(Reps.Reps["ПРИЧ АО КРАС"].Value) == convertToInteger(1)),error,Reps.Reps["ХОД АВОСТ КРАС"].Value,Reps.Reps["ДАТА АО КРАС"].Value,39,122)// по команде
				}

				//---------- Проверка пожарной сигнализации ---------------------------


				if front((convertToInteger(Reps.Reps["РАЗР2 ДП КРАС"].Value) == convertToInteger(2)),123) {//подтверждено диспетчером {
					codps = 0 // сбросить код
				aout[42]=time.Now().Unix()
				aout[43]=0
				dout[44]=1
			}

			if convertToInteger(dout[44]) == convertToInteger(1) { //выполняется {
				codps = setbitif (convertToInteger(Reps.Reps["ЯХПЖ ПЕР КРАС"].Value) == convertToInteger(2),codps,0) // появился синал пожара блока переключений
					codps = setbitif (convertToInteger(Reps.Reps["ЯХПЖ РЕД КРАС"].Value) == convertToInteger(2),codps,1) // появился синал пожара блока ред
						codps = setbitif (convertToInteger(Reps.Reps["ЯХПЖ ОПЕР КРАС"].Value) == convertToInteger(2),codps,2) // появился синал пожара блока опе
							codps = setbitif (convertToInteger(Reps.Reps["ЯХПЖ ИПР КРАС"].Value) == convertToInteger(2),codps,3) // появился синал пожара ипр

								aout[43]=time.Now().Unix()-aout[42]

								if convertToInteger(codps) == convertToInteger(15) { // успешно {
									dout[44]=2
								}

								if convertToInteger(time.Now().Unix()-aout[42]) > convertToInteger(30*60) { // время вышло 30 мин {
									dout[44]=3 // неудача
								}
							}

							aout[95]=codps

							//---------- Проверка системы контроля загазованности ---------------------------

							if front((convertToInteger(Reps.Reps["РАЗР3 ДП КРАС"].Value) == convertToInteger(2)),124) {//подтверждено диспетчером {
								codz = 0 // сбросить код
							aout[45]=time.Now().Unix()
							aout[46]=0
							dout[47]=1
						}

						if convertToInteger(dout[47]) == convertToInteger(1) { //выполняется {
							codz = setbitif (Reps.Reps["СН-2 ПЕР КРАС"].Value,codz,0) // появился синал пожара блока переключений
							codz = setbitif (Reps.Reps["СН-2 РЕД КРАС"].Value,codz,1) // появился синал пожара блока ред

									aout[46]=time.Now().Unix()-aout[45] // таймер

									if convertToInteger(codz) == convertToInteger(3) { // успешно {
										dout[47]=2
									}

									if convertToInteger(aout[46]) > convertToInteger(30*60) { // время вышло 30 мин {
										dout[47]=3 // неудача
									}
								}

								aout[96]=codz

								//----
			if convertToInteger(Reps.Reps["РАЗР5 ДП КРАС"].Value) == convertToInteger(2) {   //подтверждено диспетчером {

				// -- Ревизия фильтров
				run  =((convertToInteger(Reps.Reps["КР ЛРЕД1 КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["DP ФИЛЬТ1 КРАС"].Value) < convertToInteger(0.1))) || ((convertToInteger(Reps.Reps["КР ЛРЕД2 КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["DP ФИЛЬТ2 КРАС"].Value) < convertToInteger(0.1)))
				stop = (convertToInteger(Reps.Reps["КР ЛРЕД1 КРАС"].Value) == convertToInteger(1)) && (convertToInteger(Reps.Reps["КР ЛРЕД2 КРАС"].Value) == convertToInteger(1))
				x=testTrackRR(run,stop,51,128)//

				// -- Ревизия ПГ
				run  =(convertToInteger(Reps.Reps["КРАН БППГ КРАС"].Value) == convertToInteger(1)) && (convertToInteger(Reps.Reps["КР ПГВХ КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["КР ПГВЫХ КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["ПГ РЕЖ КРАС"].Value) == convertToInteger(0))
				stop =(convertToInteger(Reps.Reps["КРАН БППГ КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["КР ПГВХ КРАС"].Value) == convertToInteger(1)) && (convertToInteger(Reps.Reps["КР ПГВЫХ КРАС"].Value) == convertToInteger(1)) && (convertToInteger(Reps.Reps["ПГ РЕЖ КРАС"].Value) == convertToInteger(4))
				x=testTrackRR(run,stop,54,129)//

				// -- Ревизия регуляторов
				run  =((convertToInteger(Reps.Reps["КР ЛРЕД1 КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["РРЕГ1 КРАС"].Value) < convertToInteger(0.1))) || ((convertToInteger(Reps.Reps["КР ЛРЕД2 КРАС"].Value) == convertToInteger(2)) && (convertToInteger(Reps.Reps["РРЕГ2 КРАС"].Value) < convertToInteger(0.1)))
				stop = (convertToInteger(Reps.Reps["КР ЛРЕД1 КРАС"].Value) == convertToInteger(1)) && (convertToInteger(Reps.Reps["КР ЛРЕД2 КРАС"].Value) == convertToInteger(1))
				x=testTrackRR(run,stop,57,130)//

				// -- рез электропитание
				x=testTrackRR(Reps.Reps["UPS1 РЕЖ КРАС"].Value,!Reps.Reps["UPS1 РЕЖ КРАС"].Value,60,131)//
			}

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
