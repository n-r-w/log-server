# Пример сервера логов на Go, построенного по принципам Clean Architecture (Golang log server server with a clean architecture approach)

Пример сервера логов на Go, построенного по принципам Clean Architecture. Пример учебный и в проде не использовался.
Сервер работает по http rest и имеет следующие функции:
* Аутентификация
* Добавление/удаление пользователей
* Смена собственного пароля или пароля другого пользователя (только админ)
* Добавление логов
* Запрос логов по интервалу дат

Ответ на запрос логов может быть в виде:
* JSON
* JSON упакованный gzip
* Protocol Buffers упакованный gzip

Формат ответа определяется HTTP хедерами запроса
