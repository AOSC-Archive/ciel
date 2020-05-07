package systemd

import (
	"log"
	"os"
	"strconv"

	"github.com/godbus/dbus/v5"
)

var Conn *dbus.Conn

func init() {
	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		log.Panicln(err)
	}
	authMethods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}
	err = conn.Auth(authMethods)
	if err != nil {
		log.Panicln(err)
	}
	err = conn.Hello()
	if err != nil {
		log.Panicln(err)
	}
	Conn = conn
}
