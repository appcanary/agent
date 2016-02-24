package agent

import (
	"net"
	"os"
	"os/exec"

	"github.com/appcanary/agent/agent/detect"
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

// Creates a new server and syncs conf if needed
func NewServer(name string, conf *ServerConf) *Server {
	var err error
	var hostname, uname, this_ip, distro, release string

	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// syscall.Uname is only available in linux because in Darwin it's not a syscall (who knew)
	// furthermore, syscall.Uname returns a struct (syscall.Utsname) of [65]int8 -- which are *signed* integers
	// instead of convering the signed integers into bytes and processing the whole thing into a string, we're just going to call uname -a for now and upload the returned string
	cmdUname, err := exec.Command("uname", "-a").Output()
	if err != nil {
		uname = "unknown"
	} else {
		uname = string(cmdUname)
	}

	// For now we only get ipv4 ips
	// If we don't find any
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, a := range addrs {
			//If we can't find a valid ipv4 ip, it will remain "unknown"
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				this_ip = ipnet.IP.String()
				break
			}
		}
	}

	osInfo, err := detect.DetectOS()
	if err != nil {
		log.Error(err.Error())
		distro = "unknown"
		release = "unknown"
	} else {
		distro = osInfo.Distro
		release = osInfo.Release
	}

	return &Server{Name: name, Hostname: hostname, Uname: uname, Ip: this_ip, UUID: conf.UUID, Distro: distro, Release: release}
}

func (server *Server) IsNew() bool {
	return server.UUID == ""
}
