package agent

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
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

	// We now support CentOS 7.
	// This needs to be overhauled but for now:

	if centosRelease, err := ioutil.ReadFile("/etc/centos-release"); err == nil {

		distro = "centos"

		sevenMatch, err := regexp.Match("release 7", centosRelease)
		if sevenMatch && (err == nil) {
			release = "7"
		} else {
			release = "unknown"
		}
	} else {

		// We can find out distro and release on debian systems
		etcIssue, err := ioutil.ReadFile("/etc/issue")
		// if we fail reading, distro/os is unknown
		if err != nil {
			distro = "unknown"
			release = "unknown"
			// log.Error(err.Error())
		} else {
			// /etc/issue looks like Ubuntu 14.04.2 LTS \n \l
			s := strings.Split(string(etcIssue), " ")
			distro = strings.ToLower(s[0])
			release = s[1]
		}
	}

	return &Server{Name: name, Hostname: hostname, Uname: uname, Ip: this_ip, UUID: conf.UUID, Distro: distro, Release: release}
}

func (server *Server) IsNew() bool {
	return server.UUID == ""
}
