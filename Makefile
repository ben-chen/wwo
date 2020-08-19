wwo: wwo.o
	gccgo -o wwo wwo.o

wwo.o: wwo.go
	gccgo -c wwo.go
