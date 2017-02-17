package agent

import (
	"net"
	"os"
	"os/exec"

	"github.com/appcanary/agent/agent/detect"
	"github.com/appcanary/agent/conf"
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

func (s *Server) WithOSInfo() *Server {
	log := conf.FetchLog()

	// make a copy so we can cache the OS details
	server := *s

	// if these are already set, it is from config
	if server.Distro == "" && server.Release == "" {
		guess, err := detect.DetectOS()
		if err != nil {
			log.Error(err)
			server.Distro = "unknown"
			server.Release = "unknown"
		} else {
			server.Distro = guess.Distro
			server.Release = guess.Release
		}
	}

	return &server
}

// Creates a new server and syncs conf if needed
func NewServer(agentConf *conf.Conf, serverConf *conf.ServerConf) (server *Server) {
	var err error
	var hostname, uname, thisIP string

	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// syscall.Uname is only available in linux because in Darwin it's not a
	// syscall (who knew) furthermore, syscall.Uname returns a struct
	// (syscall.Utsname) of [65]int8 -- which are *signed* integers instead of
	// convering the signed integers into bytes and processing the whole thing
	// into a string, we're just going to call uname -a for now and upload the
	// returned string
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
				thisIP = ipnet.IP.String()
				break
			}
		}
	}

	server = &Server{
		Name:     agentConf.ServerName,
		Hostname: hostname,
		Uname:    uname,
		Ip:       thisIP,
		UUID:     serverConf.UUID,
	}

	// set these if they're provided by conf, otherwise we run detect every
	// time we need them (via server.WithOSInfo())
	confOSInfo := agentConf.OSInfo()
	if confOSInfo != nil {
		server.Distro = confOSInfo.Distro
		server.Release = confOSInfo.Release
	}

	return
}

func (server *Server) IsNew() bool {
	return server.UUID == ""
}

func (server *Server) IsUbuntu() bool {
	return server.WithOSInfo().Distro == "ubuntu"
}

func (server *Server) IsCentOS() bool {
	return server.WithOSInfo().Distro == "centos"
}
