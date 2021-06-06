<p align="center">
  <img src="assets/gopher.png" height="300">
  <h1 align="center">
    TCP Echo Server
    <br>
    <a href="https://coveralls.io/github/BeSoBad/goecho?branch=main" ><img alt="coverage" src="https://coveralls.io/repos/github/BeSoBad/goecho/badge.svg?branch=main" /></a>
    <a href="https://goreportcard.com/report/github.com/BeSoBad/goecho"><img alt="go-report" src="https://goreportcard.com/badge/github.com/BeSoBad/goecho" /></a>
    <a href="https://codebeat.co/projects/github-com-besobad-goecho-main" ><img alt="codebeat" src="https://codebeat.co/badges/c1790f80-124e-443b-b8b1-79659a3c1c50" /></a>
  </h1>
</p>


Asynchronous TCP server implementing [Echo Protocol](https://datatracker.ietf.org/doc/html/rfc862)

## Features
- One goroutine handles one connection.
- Gracefully stopping the server and closing all active connections.
- Ability to write and use custom MessageHandler instead of EchoHandler.

## Build and run
- `docker-compose up`

## Connect
#### Windows
- `telnet 127.0.0.1 7`
