//package main
//
//import (
//	"fmt"
//	"math/rand"
//	"regexp"
//	"time"
//)
//
//type Rep struct {
//	Value int
//}
//
//function replaceExpressions(text string, reps map[string]Rep) string {
//	re := regexp.MustCompile(`\{([^{}\n]+)\}`)
//
//	// Заменяем выражения в тексте
//	result := re.ReplaceAllStringFunc(text, function(match string) string {
//		repName := match[1 : len(match)-1] // Извлекаем имя репера из скобок
//		if rep, found := reps[repName]; found {
//			return fmt.Sprintf("Reps[\"%s\"].Value", repName)
//		}
//		return match // Если репер не найден, оставляем выражение без изменений
//	})
//
//	return result
//}
//
//function main() {
//	// Пример использования
//	text := `
//	// ... ваш текст ...
//
//	// Аварийное Закрытие ГРС со стравливанием
//	if checkPrecondSt(0) {
//		if {РЕЖИМ ГРС ДЕС} != 0 && ({ХОД АО СТ ДЕС} == 0) && ({ХОД АО СТ ДЕС} == 0) && ({РЗР АО СТ ДЕС} == 0) {
//			dout[1]=1 // ход ао
//			// пошел останов
//			SET({табл АварияДЕС}[sys_num], 1)
//			SET({звонок шк ДЕС}[sys_num], 1)
//			x=setwex({ОХР КР ДЕС}[sys_num],1,{Т ож кран ДЕС})
//			x=setwex({Вход ДЕС}[sys_num],1,{Т ож кран ДЕС})
//			x=setwex({Выход ДЕС}[sys_num],1,{Т ож кран ДЕС})
//			x=setwex({ВЫХ Д ДЕС}[sys_num],1,{Т ож кран ДЕС})
//			x=setwex({КРдоРУ ДЕС}[sys_num],1,{Т ож кран ДЕС})
//			SET({Клап котлы ДЕС}[sys_num], 1)
//
//			//set({}[sys_num],1 - отключение котлов
//			if ({Вход ДЕС} == 2) && ({Выход ДЕС} == 2) && ({ОХР КР ДЕС} == 2) && ({КРдоРУ ДЕС} == 2) // проверка закрылись ли краны
//				x=setwex({СВзаВХ ДЕС}[sys_num],0,{Т ож кран ДЕС})       // открыть свечной кран на входе
//				x=setwex({СВдоВЫХ ДЕС}[sys_num],0,{Т ож кран ДЕС})      // открыть свечной кран на выходе
//				x=setwex({СВ ОК ДЕС}[sys_num],0,{Т ож кран ДЕС})        // открыть свечной кран ОХРАН крана
//				SET({ТУ КОТЕЛ 1 ДЕС}[sys_num], 1)
//						 // выключить котел 1
//				SET({ТУ КОТЕЛ 2 ДЕС}[sys_num], 1)
//						 // выключить котел 2
//		}
//		dout[1]=2+(({ОХР КР ДЕС} == 2) && ({Вход ДЕС} == 2) && ({КРдоРУ ДЕС} == 2) && ({Выход ДЕС} == 2) && ({ВЫХ Д ДЕС} == 2) && ({СВзаВХ ДЕС} == 1) && ({СВдоВЫХ ДЕС} == 1) && ({СВ ОК ДЕС} == 1))  // ход ао
//		SET({КОМ РЕЖ3 ДЕС}[sys_num], 0)
//		// перевод в информ режим
//	}
//	}
//	// ... ваш текст ...
//	`
//
//	// Инициализируем генератор случайных чисел
//	rand.Seed(time.Now().UnixNano())
//
//	// Инициализируем реперы
//	reps := make(map[string]Rep)
//	reps["ХОД АО СТ ДЕС"] = Rep{Value: rand.Intn(2)}
//	reps["ОХР КР ДЕС"] = Rep{Value: rand.Intn(2)}
//	reps["Вход ДЕС"] = Rep{Value: rand.Intn(2)}
//	// Добавьте другие реперы по аналогии
//
//	// Заменяем выражения в тексте
//	result := replaceExpressions(text, reps)
//
//	// Выводим результат
//	fmt.Println(result)
//}
