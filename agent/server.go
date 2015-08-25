package agent

import (
	"net"
	"os"
	"os/exec"
)

type Server struct {
	Hostname     string `json:"hostname"`
	Uname        string `json:"uname"`
	Ip           string `json:"ip"`
	Name         string `json:"name"`
	UUID         string `json:"uuid,omitempty"`
	DistroString string `json:"distro-string,omitempty"`
}

// Creates a new server and syncs conf if needed
func NewServer(conf *ServerConf) *Server {
	var err error
	if conf.Hostname == "" {
		conf.Hostname, err = os.Hostname()
		if err != nil {
			conf.Hostname = "unknown"
		}
	}
	if conf.Uname == "" {
		// syscall.Uname is only available in linux because in Darwin it's not a syscall (who knew)
		// furthermore, syscall.Uname returns a struct (syscall.Utsname) of [65]int8 -- which are *signed* integers
		// instead of convering the signed integers into bytes and processing the whole thing into a string, we're just going to call uname -a for now and upload the returned string
		u, err := exec.Command("uname", "-a").Output()
		if err != nil {
			conf.Uname = "unknown"
		} else {
			conf.Uname = string(u)
		}
	}

	if conf.Ip == "" {
		// For now we only get ipv4 ips

		// If we don't find any
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			for _, a := range addrs {
				//If we can't find a valid ipv4 ip, it will remain "unknown"
				if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					conf.Ip = ipnet.IP.String()
					break
				}
			}
		}
	}

	conf.ParseDistro()

	return &Server{Hostname: conf.Hostname, Uname: conf.Uname, Ip: conf.Ip, UUID: conf.UUID, DistroString: conf.DistroString}
}

func (server *Server) IsNew() bool {
	return server.UUID == ""
}
