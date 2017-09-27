# boonfoto
Golang based photo organizer to manage assets on NAS.

To compile for arm from project base directory.

```text
CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOPATH=$PWD go build -i -o boonfoto cmd/boonfoto/*.go
```