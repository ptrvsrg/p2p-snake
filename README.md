# Узел P2P-игры "Змейка"

<p align="center">
   <a href="https://github.com/ptrvsrg/p2p-snake-node/graphs/contributors">
        <img alt="GitHub contributors" src="https://img.shields.io/github/contributors/ptrvsrg/p2p-snake-node?style=flat&label=Contributors&labelColor=222222&color=77D4FC"/>
   </a>
   <a href="https://github.com/ptrvsrg/p2p-snake-node/forks">
        <img alt="GitHub forks" src="https://img.shields.io/github/forks/ptrvsrg/p2p-snake-node?style=flat&label=Forks&labelColor=222222&color=77D4FC"/>
   </a>
   <a href="https://github.com/ptrvsrg/p2p-snake-node/stargazers">
        <img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/ptrvsrg/p2p-snake-node?style=flat&label=Stars&labelColor=222222&color=77D4FC"/>
   </a>
   <a href="https://github.com/ptrvsrg/p2p-snake-node/issues">
        <img alt="GitHub issues" src="https://img.shields.io/github/issues/ptrvsrg/p2p-snake-node?style=flat&label=Issues&labelColor=222222&color=77D4FC"/>
   </a>
   <a href="https://github.com/ptrvsrg/p2p-snake-node/pulls">
        <img alt="GitHub pull requests" src="https://img.shields.io/github/issues-pr/ptrvsrg/p2p-snake-node?style=flat&label=Pull%20Requests&labelColor=222222&color=77D4FC"/>
   </a>
</p>

Добро пожаловать в репозиторий P2P Snake Node! Это проект узла P2P-игры змейка, реализующего
взаимодействие с другими узлами предоставляющего и предоставляющего UDP API

## Технологии

- Golang 1.20
- Protocol Buffers 2
- Logrus
- Viper

## Описание

Узел состоит из нескольких компонентов:

- [Парсер командной строки](./internal/clparser)
- [Парсер конфигурационного файла](./internal/config)
- [Логгер](./internal/log)
- [Игровой движок](./internal/engine)
- [P2P узел](./internal/p2p)
- [API сервер](./internal/api)
- [Детектор копий](./internal/hub)

### Парсер командной строки

Доступны 2 флага:

- `--config` - указывает путь к конфигурационному файлу (в случае отсутствия, ожидается
  конфигурационный файл config/config.json)
- `-v` - определяет будет ли узел виден (будет ли работать детектор копий)

### Парсер конфигурационного файла

Ожидается следующий формат:

```json
{
    "p2p": {
        "delay": 1000,
        "multicast": {
            "address": "239.192.0.4",
            "port": 9192
        }
    },
    "api": {
        "public_url": "192.168.3.43:9193",
        "port": 9193,
        "timeout": 1000
    },
    "hub": {
        "multicast": {
            "address": "239.192.0.5",
            "port": 9194
        }
    }
}
```

### Логгер

Пример логов:

```text
2023-12-24 23:41:31.550  INFO --- /p2p-snake/cmd/p2p-snake/main.go:33 : to quit application press Ctrl+C 
2023-12-24 23:41:31.550  INFO --- /p2p-snake/internal/api/server.go:74 : API server is listening on :9196 
2023-12-24 23:41:31.550  INFO --- /p2p-snake/internal/p2p/peer.go:71 : P2P node is listening on multicast 239.192.0.4:9192 
2023-12-24 23:41:31.550  INFO --- /p2p-snake/internal/p2p/peer.go:79 : P2P node is listening on unicast [::]:45322 
2023-12-24 23:41:31.551  INFO --- /p2p-snake/internal/hub/sender.go:55 : hub sender running on [::]:60610 
```

### Игровой движок

- Содержит данные игры: змеек, игроков, координаты пищи
- Реализует логику игры: вычисление следующего состояния, добавление змеек и еды, проверка змеек на
  поедание пищи и на столкновение как с самой собой, так и с другими змейками

### P2P узел

Отвечает за общение с другими P2P-узлами, реализует протокол общения узлов ([подробное описание
протокола](./docs/TASK.md), [protobuf файл протокола](./protocol/p2p.proto))

### API сервер

Предоставление UDP API для клиентов, поддержание "соединения" с клиентом, делегирование запросов
клиентов p2p-узлу ([protobuf файл протокола](./protocol/api.proto))

### Детектор копий

В случае установленного флага `-v` будет отправляться сообщение на мультикаст адрес хаба, который
могут прослушивать не имеющие своего узла клиенты в поисках свободного. Как только какой-нибудь
клиент "соединяется" с узлом, сообщения перестают
отправляться ([protobuf файл протокола](./protocol/hub.proto))

## Установка и настройка

### Вручную

1. Убедитесь, что у вас установлен Golang 1.20;
2. Клонируйте репозиторий на свою локальную машину;
3. Запустите приложение с помощью команды:

   ```shell
   make build
   ./p2p-snake-node -v -config <CONFIG_PATH>
   ```
   
### Docker

1. Убедитесь, что у вас установлен Docker;
2. Запустите контейнер с помощью команды:

    ```shell
    sudo docker run \
    -d \
    -v ./config.json:/config.json \
    -e CONFIG_FILE=/config.json \
    --network host \
    --name p2p-snake-node \
    ptrvsrg/p2p-snake-node:latest
    ```

## Вклад в проект

Если вы хотите внести свой вклад в проект, вы можете следовать этим шагам:

1. Создайте форк этого репозитория.
2. Внесите необходимые изменения.
3. Создайте pull request, описывая ваши изменения.
