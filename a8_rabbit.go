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


// ГРС Красноусольский
// Галеев 07.2019
// Мелкие алгоритмы, влияющие на объекты

//
// час текущего времени сау
//



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


func oninit(Reps *SafeMap) {
	//b:=0
	//d:=0
	//prevhour:=curhour()
	dout[1]=0
	dout[2]=0
	aout[3]=0
	dout[4]=int(Reps.Reps["АВТОЗАПУСКР"].Value)      // уст извне, сохранение при перезагр
	dout[5]=0                      // при старте, считаем, кнопки не нажаты
	dout[6]=0
	dout[7]=0
	dout[8]=0
	aout[9]=0
	aout[10]=0
	dout[11]=0
	dout[12]=0
	aout[15]=0

	dout[18]=int(Reps.Reps["АВТОЗАПУСКР"].Value)
	dout[19]=int(Reps.Reps["АВТОЗАПУСКР"].Value)
	dout[21]=0
	aout[22]=0
	dout[23]=0

	aout[24]=0
	aout[25]=0
	aout[26]=0
	aout[27]=0
	time.Sleep((5*18) * time.Second)
}

func mainOutput(Reps *SafeMap) {
	// реакция на команды и кнопки перехода в режим
	//
	modes(4,&Reps.Reps["КОМ РЕЖ1 КРАС"].Value,&Reps.Reps["КОМ РЕЖ2 КРАС"].Value,&Reps.Reps["КН РЕЖ1 КРАС"].Value,&Reps.Reps["КН РЕЖ2 КРАС"].Value)
	cmdmode_in(4,18,&Reps.Reps["КОМ РЕЖ3"].Value)
	//
	// блокировка команд ТУ в режимах АРМ и ПУДП (1-команды разрешены)
	// режим грс 0-по месту, 1-пу, 2-арм
	//
	// упр кранами
	if (Reps.Reps["РЕЖИМ ГРС КРАС"].Value) == 1 {
		SET(&Reps.Reps["ТУ ОБ СПУ"].Value, 1.0)
	} else {
		SET(&Reps.Reps["ТУ ОБ СПУ"].Value, 0.0)
	}
	if (Reps.Reps["РЕЖИМ ГРС КРАС"].Value) == 2 {
		SET(&Reps.Reps["ТУ ОБ АРМ"].Value, 1.0)
	} else {
		SET(&Reps.Reps["ТУ ОБ АРМ"].Value, 0.0)
	}

	// режим неуправляемый, при изменениях сохраняем на диск
	//
	SET (&Reps.Reps["АВТОЗАПУСКР"].Value ,Reps.Reps["РЕЖИМ ГРС КРАС"].Value)
	// если режим пу дп и нет связи с дп - перевести режим в арм
	// 300сек- в modbus_s "время обрыва связи"
	//
	if (convertToInteger(Reps.Reps["РЕЖИМ ГРС КРАС"].Value) == convertToInteger(1)) && (convertToInteger(DOST(Reps.Reps["СВЯЗ ЛПУ КРАС"].Value)) == convertToInteger(1)) && (convertToInteger(Reps.Reps["СВЯЗ ЛПУ КРАС"].Value) == convertToInteger(0)) {
		dout[4]=2
	}
	// если нет связи с бус в течение 120с - выдать сигнализацию
	//
	if Reps.Reps["СВЯЗЬ С БУС1"].Reliability {
		if valTrack(Reps.Reps["СВЯЗЬ С БУС1"].Value,120,21) {dout[13]= 1 } else {dout[13]= 0}
	}
	if Reps.Reps["СВЯЗЬ С БУС2"].Reliability {
		if valTrack(Reps.Reps["СВЯЗЬ С БУС2"].Value,120,22) {dout[14]= 1} else {dout[14]= 0}
	}
	//-------------- засылка расхода газа в одоризатор -------------
	//
	if (convertToInteger(Reps.Reps["КРАН БАЙП КРАС"].Value) == convertToInteger(2)) {
		setq_periodic(16,Reps.Reps["1SF QМГН КРАС"].Value,Reps.Reps["Q ЗАМ1 КРАС"].Value,Reps.Reps["РЕЖ ОДОР1 КРАС"].Value,&Reps.Reps["QАВТ БОМ КРАС"].Value,30)
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
