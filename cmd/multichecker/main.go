// Package multichecker предоставляет настраиваемый статический анализатор для Go кода.
// Примеры использования:
//   # Проверка всего проекта
//   go run ./cmd/multichecker ./...
//
//   # Проверка конкретного пакета
//   go run ./cmd/multichecker ./cmd/server
//
// Включенные анализаторы:
//
// Стандартные анализаторы (golang.org/x/tools/go/analysis/passes):
//
//   asmdecl     - Проверка соответствия объявлений ассемблерных функций
//   assign      - Обнаружение бесполезных присваиваний
//   atomic      - Проверка правильности использования пакета sync/atomic
//   bools       - Обнаружение распространенных ошибок с булевыми операторами
//   buildtag    - Проверка корректности тегов сборки
//   cgocall     - Проверка правильности вызовов Cgo
//   composite   - Проверка инициализации композитных литералов
//   copylock    - Обнаружение копирования мьютексов и других блокирующих структур
//   errorsas    - Проверка правильности использования errors.As
//   fieldalignment - Анализ выравнивания полей в структурах для оптимизации памяти
//   framepointer - Проверка работы с указателями на фреймы
//   httpresponse - Проверка обработки HTTP ответов (закрытие тела)
//   loopclosure - Обнаружение проблем с замыканиями в циклах
//   lostcancel  - Обнаружение потерянных контекстов отмены
//   nilfunc     - Обнаружение вызовов nil-функций
//   printf      - Проверка форматных строк в Printf-функциях
//   shift       - Проверка корректности операций сдвига
//   sigchanyzer - Обнаружение неправильного использования os/signal.Notify
//   sortslice   - Проверка правильности использования сортировки слайсов
//   stdmethods  - Проверка соответствия стандартным интерфейсам (String, Error)
//   stringintconv - Обнаружение преобразований строк в целые числа
//   structtag   - Проверка тегов структур
//   testinggoroutine - Обнаружение утечки горутин в тестах
//   tests       - Проверка корректности тестов
//   unmarshal   - Проверка правильности анмаршалинга
//   unreachable - Обнаружение недостижимого кода
//   unsafeptr   - Проверка преобразований unsafe.Pointer
//   unusedresult - Обнаружение неиспользуемых возвращаемых значений
//
// Staticcheck анализаторы класса SA (staticcheck.io):
//
//   SA1xxx - Различные проверки корректности кода
//   SA2xxx - Проверки распределения памяти
//   SA3xxx - Проверки тестов и бенчмарков
//   SA4xxx - Проверки логики и потоков управления
//   SA5xxx - Проверки корректности использования API
//   SA6xxx - Проверки производительности
//   SA9xxx - Прочие проверки
//
// Дополнительные staticcheck анализаторы:
//
//   S1000    - Упрощение булевых выражений
//   ST1000   - Стиль комментариев к экспортируемым элементам
//   QF1001   - Упрощение литералов среза/массива
//
// Пользовательские анализаторы:
//
//   exitanalyzer - Запрещает прямой вызов os.Exit в функции main пакета main

package main

import (
	// "github.com/your-username/your-project/internal/analyzer/noosexit"
	"github.com/Valentin-Makurin/metrics/internal/analyzer/exitanalyzer"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	// 1. Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	// 2. Все анализаторы класса SA из staticcheck
	for _, v := range staticcheck.Analyzers {
		// Берем  анализаторы, которые начинаются на "SA"
		// анализатор S1000 из класса S1xxx
		// публичныt анализаторs ST1000 (стиль комментариев) и QF1001 (quick fix для объединения литералов)
		//
		if strings.HasPrefix(v.Analyzer.Name, "SA") ||
			v.Analyzer.Name == "ST1000" ||
			v.Analyzer.Name == "QF1001" ||
			v.Analyzer.Name == "S1000" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// 3. Кастомный анализатор
	analyzers = append(analyzers, exitanalyzer.Analyzer)

	multichecker.Main(analyzers...)
}
