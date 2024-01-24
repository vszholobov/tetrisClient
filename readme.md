Сборка с указанием хоста при компиляции:
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.addr=158.160.11.79:8080"

Запуск с созданием сессии:
./tetrisClient create

Запуск с подключением к сессии:
./tetrisClient connect 123456789