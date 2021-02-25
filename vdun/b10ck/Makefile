all: b10ck tcp_force_reset

b10ck:
	go build

tcp_force_reset: tcp_force_reset.c
	gcc -g -Wall -lnetfilter_queue -o $@ $< 

clean:
	rm -f b10ck tcp_force_reset
