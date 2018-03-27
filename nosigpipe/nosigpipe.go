//+build !darwin !go1.9

package nosigpipe

import "net"

func IgnoreSIGPIPE(c net.Conn) {
}
