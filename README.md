# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


Type: inuse_space
Time: 2025-12-02 21:06:08 MSK
Duration: 120.01s, Total samples = 3076.15kB 
Showing nodes accounting for -513.90kB, 16.71% of 3076.15kB total
      flat  flat%   sum%        cum   cum%
   -1026kB 33.35% 33.35%    -1026kB 33.35%  runtime.allocm
  512.10kB 16.65% 16.71%   512.10kB 16.65%  github.com/jackc/pgx/v5/pgtype.(*Map).buildReflectTypeToType
 -512.05kB 16.65% 33.35%  -512.05kB 16.65%  github.com/jackc/pgx/v5/pgconn/ctxwatch.(*ContextWatcher).Watch.func1
  512.05kB 16.65% 16.71%   512.05kB 16.65%  runtime.acquireSudog
         0     0% 16.71%   512.10kB 16.65%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 16.71%   512.10kB 16.65%  github.com/jackc/pgx/v5.connect
         0     0% 16.71%   512.10kB 16.65%  github.com/jackc/pgx/v5/pgtype.NewMap
         0     0% 16.71%   512.10kB 16.65%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
         0     0% 16.71%   512.10kB 16.65%  github.com/jackc/pgx/v5/pgxpool.NewWithConfig.func3
         0     0% 16.71%   512.10kB 16.65%  github.com/jackc/puddle/v2.(*Pool[go.shape.*uint8]).initResourceValue.func1
         0     0% 16.71%   512.05kB 16.65%  runtime.gcBgMarkWorker
         0     0% 16.71%   512.05kB 16.65%  runtime.gcMarkDone
         0     0% 16.71%      513kB 16.68%  runtime.gopreempt_m
         0     0% 16.71%      513kB 16.68%  runtime.goschedImpl
         0     0% 16.71%    -1026kB 33.35%  runtime.mcall
         0     0% 16.71%      513kB 16.68%  runtime.morestack
         0     0% 16.71%     -513kB 16.68%  runtime.mstart
         0     0% 16.71%     -513kB 16.68%  runtime.mstart0
         0     0% 16.71%     -513kB 16.68%  runtime.mstart1
         0     0% 16.71%    -1026kB 33.35%  runtime.newm
         0     0% 16.71%      513kB 16.68%  runtime.newstack
         0     0% 16.71%    -1026kB 33.35%  runtime.park_m
         0     0% 16.71%    -1539kB 50.03%  runtime.resetspinning
         0     0% 16.71%    -1539kB 50.03%  runtime.schedule
         0     0% 16.71%   512.05kB 16.65%  runtime.semacquire
         0     0% 16.71%   512.05kB 16.65%  runtime.semacquire1
         0     0% 16.71%    -1026kB 33.35%  runtime.startm
         0     0% 16.71%    -1026kB 33.35%  runtime.wakep
         0     0% 16.71%   512.10kB 16.65%  sync.(*Once).Do
         0     0% 16.71%   512.10kB 16.65%  sync.(*Once).doSlow