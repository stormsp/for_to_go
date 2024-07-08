package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
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

//_________________________________________________________________________________

//func initializeSafeMap(safeMap SafeMap) *SafeMap {
//	safeMap.Rep = make(map[string]Rep)
//	return safeMap
//}


func mainOutput(Reps *SafeMap) {
	//fmt.Println("получил заявку")
	//reason := checkPrecond(Reps)
	reason := 2
	if convertToInteger(reason) != 0 {

		dout := make(map[int]int)
		aout := make(map[int]int)

		dout[2] = 1 // ход ао
		dout[3] = convertToInteger(reason)
		aout[5] = int(time.Now().Unix())

		// закрыть охранный кран
		//fmt.Println(Reps.Reps["КРАН ОХР КРАС"].Value)
		//setwex(&Reps.Reps["КРАН ОХР КРАС"].Value, 1, 40)
		setwex(&Reps.Reps["КОМ РЕЖ1 КРАС"].Value, 1, 40)
		//fmt.Println("изменил значение")
		//fmt.Println(Reps.Reps["КРАН ОХР КРАС"].Value)
		//time.Sleep(18 * time.Second)
		//if convertToInteger(Reps.Reps["КРАН ОХР КРАС"].Value) != 2 {
		if convertToInteger(Reps.Reps["КОМ РЕЖ1 КРАС"].Value) != 2 {
			// закрыть входной кран
			//setwex(&Reps.Reps["КРАН ВХОД КРАС"].Value, 1, 20)
			setwex(&Reps.Reps["КОМ РЕЖ2 КРАС"].Value, 1, 20)
		}

		// закрыть байпасный кран
		//setwex(&Reps.Reps["КРАН БАЙП КРАС"].Value, 1, 20)
		setwex(&Reps.Reps["ХОД АО КРАС"].Value, 1, 20)

		// закрыть выходной
		//setwex(&Reps.Reps["КРАН ВЫХ КРАС"].Value, 1, 20)

		// подогреватель отключить
		//SET_WAIT(&Reps.Reps["ПГ УПР КРАС"].Value, 2, 20)

		// отключить одоризатор
		//SET_WAIT(&Reps.Reps["РЕЖ ОДОР1 КРАС"].Value, 0, 20)

		// Если пожар
		//if checkFire(Reps) {

			//// если закрыты: Охранный, байпасный, выходной краны
			//if (convertToInteger(Reps.Reps["КРАН ОХР КРАС"].Value) == 2) && (convertToInteger(Reps.Reps["КРАН ВЫХ КРАС"].Value) == 2) && (convertToInteger(Reps.Reps["КРАН БАЙП КРАС"].Value) == 2) {
			//	// открыть свечные краны
			//	setwex(&Reps.Reps["КР СВ НИЗ КРАС"].Value, 0, 30)
			//	setwex(&Reps.Reps["КР СВ ВЫС КРАС"].Value, 0, 30)
			//}
			//
			//// если охранный кран не закрыт, а закрыты: входной, байпасный, выходной краны
			//if (convertToInteger(Reps.Reps["КРАН ОХР КРАС"].Value) != 2) && (convertToInteger(Reps.Reps["КРАН ВХОД КРАС"].Value) == 2) && (convertToInteger(Reps.Reps["КРАН ВЫХ КРАС"].Value) == 2) {
			//	// открыть свечной кран с низ стороны
			//	setwex(&Reps.Reps["КР СВ НИЗ КРАС"].Value, 0, 30)
			//}
		//}

		// переводим грс в режим по месту
		SET(Reps.Reps["РЕЖИМ ГРС КРАС"].MEK_Address, 0)
		//time.Sleep(5 * 18 * time.Second)
		dout[1] = 0 // ком ао (возм причина)
		dout[2] = 0

		aout[6] = int(time.Now().Unix())
	}

	if front(func() int { if convertToInteger(Reps.Reps["РЕЖИМ ГРС КРАС"].Value) != 0 { return 1 } else { return 0 } }(), 9) {
		dout[3] = 0
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
