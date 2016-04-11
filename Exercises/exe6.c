#include <stdio.h>
#include <sys/socket.h>
#include <sys/select.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <stdlib.h>

void primary(int socket, int port, int state){
	struct sockaddr_in addr;

	addr.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
	addr.sin_port = htons(port);
	addr.sin_family = AF_INET;

	while (1) {
		state++;
		sendto(socket, &state, sizeof(int), 0, (struct sockaddr*)&addr, sizeof(addr));
		printf("%d\n", state);
		sleep(1);
	}
}

void backup(int socket, int port, int state){
	struct sockaddr addr;
	socklen_t fromlen;
	struct timeval timer;
	fd_set readfds;

	while (1) {
		timer.tv_sec = 2;
		timer.tv_usec = 0;

		FD_ZERO(&readfds);
		FD_SET(socket, &readfds);

		int event = select(socket+1,&readfds,0,0,&timer);

		switch (event) {
			case -1:
				break;
			case 0:
				// Primary is not broadcasting: I am optimus primary
				puts("I am primary");
				primary(socket, port, state);
				break;
			default:
				// Primary is broadcasting:
				if (FD_ISSET(socket, &readfds)) {
					recvfrom(socket, &state, sizeof(int), 0, &addr, &fromlen);
				}
		}
	}
}

int main(int argc, char**argv) {
	struct sockaddr_in addr;
	struct timeval timer; 
	int state;
	int local_port;
	int remote_port;

	local_port = atoi(argv[1]);
	remote_port = atoi(argv[2]);

	addr.sin_family = AF_INET;
	addr.sin_addr.s_addr = INADDR_ANY;
	addr.sin_port = htons(local_port);

	int sd = socket(AF_INET, SOCK_DGRAM, 0);
	bind(sd, (struct sockaddr*) &addr, sizeof(struct sockaddr_in));

	timer.tv_sec = 2;
	timer.tv_usec = 0;

	fd_set readfds;
	FD_ZERO(&readfds);
	FD_SET(sd, &readfds); 

	state = 0;
	int event = select(sd+1,&readfds,0,0,&timer);

	switch (event) {
		case -1:
			break;
		case 0:
			// Primary is not broadcasting:
			puts("I am primary");
			primary(sd, remote_port, state);
			break;
		default:
			// Primary is broadcasting:
			if (FD_ISSET(sd, &readfds)) {
				puts("I am backup.");
				backup(sd, remote_port, state);
			}
	}
	return 0;
}