AvtioTest

Цель задания – разработать приложение имплементацию in-memory Redis кеша.

## Необходимый функционал:
Клиент и сервер tcp(telnet)/REST API
Key-value хранилище строк, списков, словарей
Возможность установить TTL на каждый ключ
Реализовать операторы: GET, SET, DEL, KEYS
Реализовать покрытие несколькими тестами функционала

## Дополнительно(необязательно):
```
Реализовать операторы: HGET, HSET, LGET, LSET
Реализовать сохранение на диск
Масштабирование (на серверной или на клиентское стороне)
Авторизация
Нагрузочные тесты
```

## REST HTTP приложение
Для запуска HTTP сервера нужно выполнить метод Run класса App.
Перед запуском сервера нужно проинициализировать инстанс App методом
Initialize(), который принимает на вход те же параметры, что и NewCache.
При инициализации приложение создаст кэши и http handler'ы
Метод Run принимает:
```
addr         string - host+port, которые будет слушать http сервер
readTimeout  int    - timeout чтения в секундах
writeTimeout int    - timeout записи в секундах
```
В качестве роутера используется gorilla/mux
## REST API
REST API принимает и возвращает данные в формате JSON

При отправке TTL передаваемое значение должно быть строкой, правильно
воспринимаемой методом time.ParseDuration()

При успешном запросе возвращается HTTP код 200, при ошибке на стороне
приложения - 400

| Метод                 | Глагол | Url          | Body                                                         | Пример успешного ответаa                                                                | Пример ошибки                                                    |
|-----------------------|--------|--------------|--------------------------------------------------------------|-----------------------------------------------------------------------------------------|------------------------------------------------------------------|
| Keys                  | GET    | /            | --                                                           | ["string","map","my_key"]                                                               | --                                                               |
| Get                   | GET    | /key         | --                                                           | {"type": 1,"data": [1,"string",{"map": "of_something"},0.2,null,["nested","list",42,]]} | {"error": "key not found"}                                       |
| Get at index          | GET    | /key/index   | --                                                           | {"inner": {"one_more": {"key": "value"}}}                                               | {"error": "cant get item at index"}                              |
| Remove                | DELETE | /key         | --                                                           | "OK"                                                                                    | --                                                               |
| Set с ttl по умолчнию | POST   | /key         | {"a":42,"list":[1,{"hello":"world"}],"something":"anything"} | {"type":2,"data":{"a":42,"list":[1,{"hello":"world"}],"something":"anything"}}          | {"error":"invalid character 'a' looking for beginning of value"} |
| Set с ttl             | POST   | /key?ttl=10s | {"a":42,"list":[1,{"hello":"world"}],"something":"anything"} | {"type":2,"data":{"a":42,"list":[1,{"hello":"world"}],"something":"anything"}}          | {"error":"Malformed duration"}                                   |

## REST HTTP client
Реализует интерфейс Cache. Создается методом NewClient, который требует url сервера
REST API, таймаут соединения и логин/пароль для базовой авторизации (если она нужна)
## Развертывание
```
go get -u github.com/shpaktakur1/TestAvito
```
На примере main
1. Выполнить go run main.go
2. Использовать REST API на localhost:8080




