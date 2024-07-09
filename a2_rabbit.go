package main

import (
	PKG "AlgorithmsRabbit/connections"
	"sync"
	"time"
	"fmt"
)

var aout [100]float32
var dout [100]float32
var x int

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
	PKG.CONNECTRABBITMIB = "amqp://admin:admin@192.168.1.102:5672/"
	PKG.NameAlg = "ButtonALG"
	//Объявление входных и выходных массивов
	PKG.DeclareArrays()
	//Подключаемся к RabbitMQ
	PKG.DeclareRabbit()
	//Запрашиваем и отправляем данные
	go PKG.ConsumeFromRabbitMq(&PKG.InputMap)
	go PKG.SendToRabbitMQ(&PKG.OutputMap)
	for {
		//Если данные получены, начинаем алгоритм
		if PKG.ConnectToRabit {
			for {
				mainOutput()
				time.Sleep(200 * time.Millisecond)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
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


//#include "eval.lib\valtrack.evl"
//#include "eval.lib\set.evl"
//#include "eval.libront.evl"


//----------- Условия выполнения аварийного останова -----------------------------
// по команде с экрана АРМ или от Диспетчера
// При нажатой более 4 секунд кнопке(механической)- только при аварийной ситуации:
// - Аварийно-высокое давление
// - Пожар в операторной
// - Пожар в блоке переключения (при наличии пож.сигнализации)
// - Пожар в блоке одоризации (при наличии пож.сигнализации)
//-------------------------------------------------------------------------------
func checkFire() float32 {
	var x float32
	x = 0
	x=x + val("ПОЖАР ОПЕ КРАС") //<Пожар в операторной>.
	x=x + val("ПОЖАР ПЕР КРАС") //<Пожар в блоке переключения>.
	//x=x || val("ПОЖАР ОДО КРАС") //<Пожар в блоке одоризации>.
	return(x)
}


func checkPrecond() float32 {
	var x float32
	x = 0
	if (val("РЕЖИМ ГРС КРАС")) != (0) {
		x=x+val("КОМ АО КРАС")  //1 команда - без условий
		if ((valTrack(val("КН АВОСТ КРАС"),4,8)) == 1) {    // кнопка - только при аварийной ситуации {
			x=x + checkFire()        //2 Пожар
			x=x+3*val("РВЫХ123АВ КРАС")    //3 Аварийно-высокое давление
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
	//aout[5]=TRUE(val("ДАТА АО КРАС"))
	aout[5]=1
	//aout[6]=TRUE(val("ДАТА ЗАО КРАС"))
	aout[6]=1

	// ждем первого опроса модулей
	time.Sleep((10) * time.Second)
	return nil
}

func mainOutput() {
	x += 1
	reason:=checkPrecond()
	if x % 5 == 0 {
		PKG.UpdateVal("КОМ РЕЖ3", 0, true)
		fmt.Println("апдейтнул")
	}
	if (reason) != (0) {
		fmt.Print("попал")

		dout[2]=1	// ход ао
		dout[3]=(reason)
		aout[5]=float32(time.Now().Unix())

		// закрыть охранный кран
		setwex("КРАН ОХР КРАС",1,40)

		time.Sleep((1) * time.Second)
		if val("КРАН ОХР КРАС") != (2) {
			// закрыть входной кран
			setwex("КРАН ВХОД КРАС",1,20)
		}

		// закрыть байпасный кран
		setwex("КРАН БАЙП КРАС",1,20)

		// закрыть выходной
		setwex("КРАН ВЫХ КРАС",1,20)

		// подогреватель отключить
		set_wait("ПГ УПР КРАС",2,20)

		// отключить одоризатор
		set_wait("РЕЖ ОДОР1 КРАС",0,20)

		// Если пожар
		if checkFire() > 0{

			// если закрыты : Охранный, байпасный, выходной краны
			if ((val("КРАН ОХР КРАС")) == (2)) && ((val("КРАН ВЫХ КРАС")) == (2)) && ((val("КРАН БАЙП КРАС")) == (2)) {
				// открыть свечные краны
				setwex("КР СВ НИЗ КРАС",0,30)
				setwex("КР СВ ВЫС КРАС",0,30)
			}

			// если охранный кран не закрыт, а закрыты: входной, байпасный, выходной краны
			if (((val("КРАН ОХР КРАС")) != (2)) && ((val("КРАН ВХОД КРАС")) == (2))) && ((val("КРАН ВЫХ КРАС")) == (2)) {
				// открыть свечной кран с низ стороны
				setwex("КР СВ НИЗ КРАС",0,30)
			}
		}

		// переводим грс в режим по месту
		PKG.UpdateVal("КОМ РЕЖ3", 0, true)
		//fmt.Print("апдейт КОМ РЕЖ3")
		time.Sleep((5) * time.Second)
		dout[1]=0	// ком ао (возм причина)
		dout[2]=0

		aout[6]=float32(time.Now().Unix())
	}
	
	if front("РЕЖИМ ГРС КРАС",0)!=9 {
		dout[3]=0
	}
	fmt.Println("конец мейн")

}