package model

type VMS []VM

type VM struct {
	Name string `json:"name"`
	State string `json:"state"`
}

type VMDetail struct {
	NAME string `json:"name"`
	CPU string `json:"cpu"`
	MEMORY string `json:"mem"`
	IPADDR string `json:"ip"`
	NETMASK string `json:"mask"`
	GATEWAY string `json:"gateway"`
	PhyIP string `json:"pyhIp"`
	IMAGE string `json:"image, omitempty"`
}

type VMC struct {
	ID uint `json:"id"`
	NAME string `json:"name"`
	UUID string `json:"uuid"`
	CPU int `json:"cpu"`
	MEMORY string `json:"memory"`
	STATUS string `json:"status"`
	NETWORK `json:"network"`
}

type NETWORK struct {
	IP string `json:"ip"`
	NETMASK string `json:"netmask"`
	GATEWAY string `json:"gateway"`
	DNS []string `json:"dns"`
}

type OP struct {
	OPERATOR string `json:"operator"`
}