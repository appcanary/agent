package models

type Server struct {
	Hostname string `json:"hostname"`
	Uname    string `json:"name"`
	Ip       string `json:"ip"`
	UUID     string `json:"uuid,omitempty"`
}

func ThisServer(uuid string) *Server {
	return &Server{Hostname: "jupiter", Uname: "Darwin maryanne.lan 13.4.0 Darwin Kernel Version 13.4.0: Wed Dec 17 19:05:52 PST 2014; root:xnu-2422.115.10~1/RELEASE_X86_64 x86_64 i386 MacBook5,1 Darwin", Ip: "192.168.1.1", UUID: uuid}
}

func (self *Server) IsNew() bool {
	return self.UUID == ""
}
