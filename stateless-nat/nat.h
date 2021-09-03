#ifndef _NAT_H_
#define _NAT_H_

#include <stdint.h>

typedef struct {
    uint32_t ipv4;
    uint16_t port;
} __attribute__((__packed__)) bind_t;

#endif
