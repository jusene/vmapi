package model

type Err struct {
	Error int `json:"error"`
	Message interface{} `json:"message"`
}

type Res struct {
	Error int `json:"error"`
	Message interface{} `json:"message"`
}