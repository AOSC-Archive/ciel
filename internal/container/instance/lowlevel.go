package instance

/*
#include <termios.h>
#include <fcntl.h>
#include <stdio.h>

struct termios tty_state;
int stdin_flags;

void save_stdin_attr() {
	stdin_flags = fcntl(0, F_GETFL, 0);
}

void recover_stdin_attr() {
	if (fcntl(0, F_GETFL, 0) != stdin_flags) {
		fcntl(0, F_SETFL, stdin_flags);
		putchar('\n');
	}
}

void save_tty_attr() {
	tcgetattr(0, &tty_state);
}

void recover_tty_attr() {
	tcsetattr(0, TCSANOW, &tty_state);
}

void save_terminal_attr() {
	save_tty_attr();
	save_stdin_attr();
}

void recover_terminal_attr() {
	recover_tty_attr();
	recover_stdin_attr();
}
*/
import "C"

func init() {
	C.save_terminal_attr()
}

func RecoverTerminalAttr() {
	C.recover_terminal_attr()
}
