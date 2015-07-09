package models

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
)

type Server struct {
	Hostname string `json:"hostname"`
	Uname    string `json:"uname"`
	Ip       string `json:"ip"`
	Name     string `json:"name"`
	UUID     string `json:"uuid,omitempty"`
	Distro   string `json:"distro,omitempty"`
	Release  string `json:"release,omitempty"`
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

	var distro string
	var release string
	// We can find out distro and release on debian systems
	etcRelease, err := ioutil.ReadFile("/etc/release")
	// if we fail reading, distro/os is unknown
	if err != nil {
		distro = "unknown"
		release = "unknown"
		log.Error(err.Error())
	} else {
		// /etc/release looks like Ubuntu 14.04.2 LTS \n \l
		s := strings.Split(string(etcRelease), " ")
		distro = strings.ToLower(s[0])
		release = s[1]
	}

	return &Server{Hostname: hostname, Uname: uname, Ip: ip, UUID: uuid, Distro: distro, Release: release, Name: "why is this here?"}
}

func (server *Server) IsNew() bool {
	return server.UUID == ""
}
