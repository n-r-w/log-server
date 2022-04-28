# Пример сервера логов на Go, построенного по принципам Clean Architecture (Golang log server with a clean architecture approach)

Пример сервера логов на Go, построенного по принципам Clean Architecture. Пример учебный и в проде не использовался.
Сервер работает по http rest, в качестве БД используется postgresql (инициализация БД в каталоге migration). Вместо postgres можно использовать тестовое хранилище в оперативной памяти.

Имеет следующие функции:
* Аутентификация
* Добавление/удаление пользователей
* Смена собственного пароля или пароля другого пользователя (только админ)
* Добавление логов
* Запрос логов по интервалу дат

Ответ на запрос логов может быть в виде:
* JSON
* JSON упакованный gzip
* JSON упакованный deflate
* Protocol Buffers упакованный gzip

Формат ответа определяется HTTP хедерами запроса

Чего тут нет:
* DTO (data transfer object) как таковые отсутствуют и структуры данных домена ползают по всем слоям. В таком простом проекте не было смысло делать маппинг DTO, да и в реальном проекте он нужен в тот момент, когда структуры данных слоев начинают расходиться. Нет смысла раньше времени делать простое сложным.
* Тест кейсы имеют довольно слабое покрытие

Планируется добавить:
* GRPС на слое presentation. Сейчас там только HTTP REST

Нагрузочное тестирование проводилось C++ клиентом: https://github.com/n-r-w/loglib

## Примеры запросов
Логин (надо сохранить полученный в ответе куки logserver для следующих запросов)

    curl --location --request POST 'http://localhost:8080/login' \
    --header 'Content-Type: application/json' \    
    --data-raw '{"login": "admin", "password": "123"}'

Получить логи за период

    curl --location --request GET 'http://localhost:8080/private/records' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODY2MHxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXw8B2eSdqLJfQJEhsrqGnuCrf5l2_ofcwCgA0Zn0sUErg==' \
    --data-raw '{"timeFrom": "2021-04-23T14:37:36.546Z","timeTo": "2022-04-23T18:25:43.511Z"}'

Получить список пользователей

    curl --location --request GET 'http://localhost:8080/private/users' \
    --header 'Cookie: logserver=MTY1MTE0ODcwNHxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXwuhL1Tz50lNOOEU6N_k2oWo6wJd1ripsKVaKIJ6XxEIw=='

Состояние аутентификации

    curl --location --request GET 'http://localhost:8080/private/whoami' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg=='

Добавить логи

    curl --location --request POST 'http://localhost:8080/private/add-log' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg==' \
    --data-raw '[{"logTime": "2020-04-23T18:25:43.511Z", "level": 4, "message1": "ошибка №2"}]'

Добавить пользователя

    curl --location --request POST 'http://localhost:8080/private/add-user' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg==' \
    --data-raw '{"login": "user11","name": "user11!!!","password": "1111"}'

Сменить пароль

    curl --location --request PUT 'http://localhost:8080/private/change-password' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg==' \
    --data-raw '{"login": "user10", "password": "1111" }'

Завершить сессию

    curl --location --request DELETE 'http://localhost:8080/close' \
    --header 'Cookie: logserver=MTY1MTE0ODkwN3xEdi1CQkFFQ180SUFBUkFCRUFBQUJQLUNBQUE9fJ6mswXN2vd3W_DpWOh7AsKYuaJiF2hd10JEUZOkKUTb'
