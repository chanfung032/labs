// +build ignore

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <time.h>
#include <libgen.h>
#include <netinet/ip.h>
#include <netinet/tcp.h>
#include <linux/netfilter.h>
#include <libnetfilter_queue/libnetfilter_queue.h>
#include <asm/byteorder.h>

#ifdef __LITTLE_ENDIAN
#define IPQUAD(addr) \
  ((unsigned char *)&addr)[0],                  \
    ((unsigned char *)&addr)[1],                \
    ((unsigned char *)&addr)[2],                \
    ((unsigned char *)&addr)[3]
#else
#define IPQUAD(addr)                            \
  ((unsigned char *)&addr)[3],                  \
    ((unsigned char *)&addr)[2],                \
    ((unsigned char *)&addr)[1],                \
    ((unsigned char *)&addr)[0]
#endif

typedef struct _pseudo_header {
    uint32_t source_ip;
    uint32_t dest_ip;
    uint8_t  reserve;
    uint8_t  protocol;
    uint16_t tcp_length;
} psHdr;

static uint   CRC32[256];
static char   init = 0;

static void init_table()
{
    int   i,j;
    uint   crc;
    for(i = 0;i < 256;i++)
    {
         crc = i;
        for(j = 0;j < 8;j++)
        {
            if(crc & 1)
            {
                 crc = (crc >> 1) ^ 0xEDB88320;
            }
            else
            {
                 crc = crc >> 1;
            }
        }
         CRC32[i] = crc;
    }
}

uint crc32( unsigned char *buf, int len)
{
    uint ret = 0xFFFFFFFF;
    int   i;
    if( !init )
    {
         init_table();
         init = 1;
    }
    for(i = 0; i < len;i++)
    {
         ret = CRC32[((ret & 0xFF) ^ buf[i])] ^ (ret >> 8);
    }
     ret = ~ret;
    return ret;
}

uint16_t checksum_calc_asm(uint16_t*, uint16_t);
// http://locklessinc.com/articles/tcp_checksum/
__asm__("; \
	.globl checksum_calc_asm; \
	.type checksum_calc_asm,@function; \
	.align 16; \
checksum_calc_asm:; \
	xor %eax, %eax; \
	test $7, %edi; \
	jne 5f; \
0:	mov %esi, %edx; \
	mov %esi, %ecx; \
	shr $3, %edx; \
	je 2f; \
	shr $6, %ecx; \
	and $7, %edx; \
	je 1f; \
	add $1, %ecx; \
	lea -64(%rdi,%rdx,8), %rdi; \
	neg %edx; \
	and $7, %edx; \
	lea 1f-1(,%rdx,4), %rdx; \
	jmp *%rdx; \
.align 16; \
1:	adc (%rdi), %rax; \
	adc 8(%rdi), %rax; \
	adc 16(%rdi), %rax; \
	adc 24(%rdi), %rax; \
	adc 32(%rdi), %rax; \
	adc 40(%rdi), %rax; \
	adc 48(%rdi), %rax; \
	adc 56(%rdi), %rax; \
	lea 64(%rdi), %rdi; \
	dec %ecx; \
	jne 1b; \
	adc $0, %rax; \
2:	and $7, %esi; \
	jne  4f; \
3:	mov %eax, %edx; \
	shr $32, %rax; \
	add %edx, %eax; \
	adc $0, %eax; \
	mov %eax, %edx; \
	shr $16, %eax; \
	add %dx, %ax; \
	adc $0, %ax; \
	not %ax; \
	retq; \
4:	lea (,%esi,8),%ecx; \
	mov $-1, %rdx; \
	neg %ecx; \
	shr %cl, %rdx; \
	and (%rdi), %rdx; \
	add %rdx, %rax; \
	adc $0, %rax; \
	jmp 3b; \
5:	test $1, %edi; \
	jne 9f; \
	mov %edi, %ecx; \
	neg %ecx; \
	and $0x7, %ecx; \
	cmp %ecx, %esi; \
	cmovl %esi, %ecx; \
	sub %ecx, %esi; \
	test $4, %ecx; \
	je  6f; \
	movl (%rdi), %edx; \
	add %rdx, %rax; \
	adc $0, %rax; \
	add $4, %rdi; \
6:	test $2, %ecx; \
	je 7f; \
	movzwq (%rdi), %rdx; \
	add %rdx, %rax; \
	adc $0, %rax; \
	add $2, %rdi; \
7:	test $1, %ecx; \
	je 8f; \
	movzbq (%rdi), %rdx; \
	add %rdx, %rax; \
	adc $0, %rax; \
	add $1, %rdi; \
8:	test %esi, %esi; \
	je 3b; \
	jmp 0b; \
9:	movzbw (%rdi), %r8w; \
	inc %rdi; \
	dec %esi; \
	je 10f; \
	call checksum_calc_asm; \
	not %ax; \
	xchg %al, %ah; \
	add %r8w, %ax; \
	adc $0, %ax; \
	not %ax; \
	retq; \
10:	mov %r8w, %ax; \
	not %ax; \
	retq; \
.size checksum_calc_asm, .-checksum_calc_asm; \
");

static int callback(struct nfq_q_handle *qh, struct nfgenmsg *nfmsg,
    struct nfq_data *nfa, void *_data)
{
    uint32_t id;
    size_t len;
    unsigned char *data;
    struct iphdr *iph;
    struct tcphdr *tcph;
    struct nfqnl_msg_packet_hdr *ph;
    int payload_length;
    int status = 0;
    uint8_t sbuf[1536];

    ph = nfq_get_msg_packet_hdr(nfa);
    if (ph) {
        id = ntohl(ph->packet_id);
        len = nfq_get_payload(nfa, &data);
        if (len > 0 && data) {
            iph = (struct iphdr*)data;
            tcph = (struct tcphdr*)(data + (iph->ihl << 2));
            payload_length = (uint32_t)ntohs(iph->tot_len) - ((iph->ihl << 2) + (tcph->doff << 2));

            if ((tcph->psh || tcph->fin || tcph->ack) && iph->protocol == IPPROTO_TCP) {
                tcph->rst = 1;
                tcph->fin = 0;
                tcph->psh = 0;
                tcph->syn = 0;
                tcph->ack = 0;
                tcph->urg = 0;

                iph->tot_len = htons(ntohs(iph->tot_len) - (uint16_t)payload_length);
                iph->check = 0;
                iph->check = checksum_calc_asm((uint16_t*)iph, iph->ihl << 2);

                tcph->check = 0;
                psHdr psh;
                int tcp_len;
                psh.source_ip = iph->saddr;
                psh.dest_ip = iph->daddr;
                psh.reserve = 0;
                psh.protocol = IPPROTO_TCP;
                tcp_len = ntohs(iph->tot_len) - (uint16_t)(iph->ihl << 2);
                psh.tcp_length =  htons(tcp_len);
                memcpy(sbuf, &psh, sizeof(psHdr));
                memcpy(sbuf + sizeof(psHdr), (unsigned char*)tcph, tcp_len);
                tcph->check = checksum_calc_asm((uint16_t*)sbuf, sizeof(psHdr) + tcp_len);
                status =  nfq_set_verdict(qh, id, NF_ACCEPT, len - payload_length, data);
            } else if (tcph->rst) {
                status = nfq_set_verdict(qh, id, NF_ACCEPT, 0, NULL);
            } else {
                status = nfq_set_verdict(qh, id, NF_DROP, 0, NULL);
            }
        }
    }
    return status;
}

int main(int argc, char **argv)
{    
    struct nfq_handle *h;
    struct nfq_q_handle *qh;
    int fd;
    int rv;
    int queueNum;
    char buf[4096] __attribute__((aligned));

    if(argc != 2) {
        fprintf(stderr, "usage: %s queueNum\n", basename(argv[0]));
        return -1;
    }
    queueNum = atoi(argv[1]);

    if(daemon(1, 1) != 0) {
        perror("daemon()");
        return -1;
    }

    fprintf(stderr, "opening library handle\n");
    h = nfq_open();
    if (!h) {
        fprintf(stderr, "error during nfq_open()\n");
        return -1;
    }

    fprintf(stderr, "unbinding existing nf_queue handler for AF_INET (if any)\n");
    if (nfq_unbind_pf(h, AF_INET) < 0) {
        fprintf(stderr, "error during nfq_unbind_pf()\n");
        return -1;
    }

    fprintf(stderr, "binding nfnetlink_queue as nf_queue handler for AF_INET\n");
    if (nfq_bind_pf(h, AF_INET) < 0) {
        fprintf(stderr, "error during nfq_bind_pf()\n");
        return -1;
    }

    fprintf(stderr, "binding this socket to queue '%d'\n", queueNum);
    qh = nfq_create_queue(h, queueNum, &callback, NULL);
    if (!qh) {
        fprintf(stderr, "error during nfq_create_queue()\n");
        return -1;
    }

    fprintf(stderr, "setting copy_packet mode\n");
    if (nfq_set_mode(qh, NFQNL_COPY_PACKET, 0xffff) < 0) {
        fprintf(stderr, "can't set packet_copy mode\n");
        return -1;
    }

    fd = nfq_fd(h);

    while ((rv = recv(fd, buf, sizeof(buf), 0)) && rv >= 0) {
        nfq_handle_packet(h, buf, rv);
    }

    fprintf(stderr, "unbinding from queue '%d'\n", queueNum);
    nfq_destroy_queue(qh);

#ifdef INSANE
    /* normally, applications SHOULD NOT issue this command, since
     * it detaches other programs/sockets from AF_INET, too ! */
    fprintf(stderr, "unbinding from AF_INET\n");
    nfq_unbind_pf(h, AF_INET);
#endif

    fprintf(stderr, "closing library handle\n");
    nfq_close(h);

    return 0;
}
