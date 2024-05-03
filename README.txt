1) Run one of these commands depending on operating system and architecture,
if the executables are not in the bin folder

LINUX
GOOS=linux GOARCH=amd64 go build -o bin/app-amd64.exe main.go

MAC (m1)
GOOS=darwin GOARCH=arm64 go build -o bin/app-arm64-darwin main.go

------------------------------------------------------------------------------------------------------------

2) After one of the executable files are in the bin, redirect
stdin and stdout with the desired input.txt and output.txt,

*if the name is different that input.txt than replace the input.txt in between the < > with the right name

Linux: .\bin/app-amd64.exe < input.txt > output.txt
Mac (m1): ./bin/app-arm64-darwin < input.txt > output.txt                                                                                                                         ─╯

------------------------------------------------------------------------------------------------------------
File descriptions: The main.go contains all my functions and the presentation shell.