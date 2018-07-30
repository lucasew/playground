#include <sys/socket.h>
#include <string.h>
#include "error.h"
#include <netinet/in.h>
#include <unistd.h>
#include "ikkp_server.h"
#include "strings.h"

// If these three functions return -1 something bad had happened, it is like if err != nil {} of golang
char *header = "Internet Knock-Knock Protocol Server\r\nVersion 1.0\r\nKnock! Knock!\r\n> ";

int ikkp_listen(int port) {
    int listener_d = socket(PF_INET, SOCK_STREAM, 0);
    if (listener_d == -1) 
        return listener_d;
    struct sockaddr_in name;
    name.sin_family = PF_INET;
    name.sin_port = (in_port_t)htons(port);
    name.sin_addr.s_addr = htonl(INADDR_ANY);
    int c = bind(listener_d, (struct sockaddr*) &name, sizeof(name));
    if (c == -1)
        return c;
    int reuse = 1;
    if(setsockopt(listener_d, SOL_SOCKET, SO_REUSEADDR, (char*)&reuse, sizeof(int)) == -1)
        return -1;
    listen(listener_d, 10); // simultaneous connections
    return listener_d;
}

int ikkp_accept(int listener_d) {
    struct sockaddr_storage client_addr;
    unsigned int address_size = sizeof(client_addr);
    int connect_d = accept(listener_d, (struct sockaddr*)&client_addr, &address_size); // connect_d
    if (ikkp_send(connect_d, header) == -1) // Firstly send header
        return -1;
    return connect_d;
}

int ikkp_send(int connect_d, char *message) {
    return send(connect_d, message, strlen(message), 0);
}

int ikkp_recv(int connect_d, void *buf, size_t len) {
    return recv(connect_d, buf, len, 0);
}

int ikkp_close(int connect_d) {
    return close(connect_d);
}
