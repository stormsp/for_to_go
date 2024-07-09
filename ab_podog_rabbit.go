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
// Отсечка подогревателя

//#include "eval.lib\set.evl"

//
func checkPrecond(Reps *SafeMap) bool {
	var x bool
	//x = 0
	if convertToInteger(Reps.Reps["РЕЖИМ ГРС КРАС"].Value) != convertToInteger(0) {
		x=x || (convertToInteger(Reps.Reps["ПГ ПРОР КРАС"].Value) == convertToInteger(1))
		x=x || (convertToInteger(Reps.Reps["КОМ ОТКПГ КРАС"].Value) == convertToInteger(1))
	}
	return(x)
}

func oninit(t any) any {
	dout[1]=0 // команда
	// ждем первого опроса модулей
	time.Sleep((10*18) * time.Second)
	return nil
}

func mainOutput(Reps *SafeMap) {
	if checkPrecond(Reps) {

		// открыть кран на байпасе подогревателя
		x=setwex(&Reps.Reps["КРАН БППГ КРАС"].Value,0,30)

		// работаем дальше если кран открылся
		if (convertToInteger(Reps.Reps["КРАН БППГ КРАС"].Value) == convertToInteger(1)) && DOST(Reps.Reps["КР ПГВХ КРАС"].Value) && DOST(Reps.Reps["КР ПГВЫХ КРАС"].Value) {

			// входной кран подогревателя закрыть
			x=setwex(&Reps.Reps["КР ПГВХ КРАС"].Value,1,30)

			// выходной кран подогревателя закрыть
			x=setwex(&Reps.Reps["КР ПГВЫХ КРАС"].Value,1,30)
			// подогреватель отключить
			x=SET_WAIT(&Reps.Reps["ПГ УПР КРАС"].Value,2,20)
		}

		dout[1]=0 // команда снята
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
