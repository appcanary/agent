package models

import (
	"net"
	"os"
	"os/exec"
)

type Server struct {
	Hostname string `json:"hostname"`
	Uname    string `json:"uname"`
	Ip       string `json:"ip"`
	Name     string `json:"name"`
	UUID     string `json:"uuid,omitempty"`
}

func ThisServer(uuid string) *Server {

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// syscall.Uname is only availabl in linux because in Darwin it's not a syscall (who knew)
	// furthermore, syscall.Uname returns a struct (syscall.Utsname) of [65]int8 -- which are *signed* integers
	// instead of convering the signed integers into bytes and processing the whole thing into a string, we're just going to call uname -a for now and upload the returned string
	u, err := exec.Command("uname", "-a").Output()
	var uname string
	if err != nil {
		uname = "unknown"
	} else {
		uname = string(u)
	}

	// For now we only get ipv4 ips

	// If we don't find any
	ip := "unknown"
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, a := range addrs {
			//If we can't find a valid ipv4 ip, it will remain "unknown"
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}

	return &Server{Hostname: hostname, Uname: uname, Ip: ip, UUID: uuid, Name: ""}
}

func (server *Server) IsNew() bool {
	return server.UUID == ""
}
