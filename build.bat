go generate
go build -ldflags "-H=windowsgui" -o GoTalk.rawbuild.exe
go build -ldflags "-H=windowsgui -s -w" -o GoTalk.stripbuild.exe