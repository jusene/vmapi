package model

type Err struct {
	Error int `json:"error"`
	Message interface{} `json:"message"`
}

type Res struct {
	Code int `json:"code"`
	Message interface{} `json:"message"`
}