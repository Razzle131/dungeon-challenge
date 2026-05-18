# dungeon-challenge
______
## Table of Contents
* [About](#about)
* [Clone repo](#clone-repo)
* [Technologies](#technologies)
* [Quick start](#quick-start)
* [Additions](#additions)

## About
This is implementation of a dungeon text game

## Clone repo
You can clone this repo to observe source code and run tests localy using this command (golang version 1.25.3 or higher):
```
git clone https://github.com/Razzle131/dungeon-challenge.git
```

## Technologies
Project is created with:
* Golang version: 1.25.3
* testify

## Quick start
### Build from source
* Ensure that Go is installed on your machine and it`s version is equal or higther than 1.25.3
```
go version
```
* Clone repo
```
git clone https://github.com/Razzle131/dungeon-challenge.git
```
* Install dependecies to run tests
```
go mod tidy
```
* Run unit tests:
```
go test ./... -cover -short
```
or
```
make unit
```
* Run programm
```
go run main.go
```
* You can use Makefile shortcut commands for running/testing/building docker image/starting docker container. For more info inspect Makefile
* You can specify events and config files by setting flags. Example:
```
go run main.go -config config.json -events events
```
______
### Docker
This app could be builded and executed using docker:
* Clone repo
```
git clone https://github.com/Razzle131/ComputerClub.git
```
* use this command to build docker image  
```
docker build -t IMAGE-NAME .
```
replace the IMAGE-NAME with the name you want

or (configure docker image name in Makefile)
```
make docker-build
```
* run docker container:
```
docker run IMAGE-NAME
```
or (configure docker image name in Makefile)
```
make docker-run
```
______
______
## Additions
### Testing
* Business logic was covered for 60% by unit tests.
* You can run ```make tests```, which runs all files in testdata folder and prints if output of programm and expected output in file differs.
