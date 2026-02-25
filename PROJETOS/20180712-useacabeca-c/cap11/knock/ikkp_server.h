#ifndef ikkp_server
// Lets define the protocol version :p
#define ikkp_server 1

// Listens for connections
int ikkp_listen(int port);

// Accepts new connections
int ikkp_accept(int listener_d);

// Sends data through the socket
int ikkp_send(int connect_d, char *message);

// Receives data through the socket
int ikkp_recv(int connect_d, void *buf, size_t len);

// Closes a opened connection
int ikkp_close(int connect_d);

#endif
