#define _GNU_SOURCE
#include <fcntl.h>
#include <unistd.h>
#include <stdio.h>

int main(int argc, char *argv[]) {
    int pfd[2];
    if (pipe(pfd) != 0) {
        perror("pipe() failed");
    }

    int infd = open(__FILE__, O_RDONLY);
    if (infd == -1) {
        perror("open input file failed");
    }

    ssize_t bytes;
    bytes = splice(infd, NULL, pfd[1], NULL, 4096, SPLICE_F_MOVE);
    printf("splice infd -> pipe bytes: %d\n", bytes);

    close(infd);

    int outfd = open(__FILE__ ".1", O_WRONLY|O_CREAT);
    if (outfd == -1) {
        perror("open output file failed");
    }
    bytes = splice(pfd[0], NULL, outfd, NULL, bytes, SPLICE_F_MOVE);
    printf("splice pipe -> outfd bytes: %d\n", bytes);

    close(outfd);
}
