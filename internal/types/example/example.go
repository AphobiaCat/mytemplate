package example

type GetExampleRequest struct {
	Name string `json:"name"`
	Ip   string `header:"ip"`
	User string `header:"user"`
	Msg  string `header:"msg"`
}

type PostExampleRequest struct {
	Name string `json:"name"`
}

type GetExampleResponse struct {
	Message string `json:"message"`
}

type PostExampleResponse struct {
	Message string `json:"message"`
}

type MidRequest struct {
	User string `header:"user"`
	Msg  string `header:"msg"`
}

type MidResponse struct {
	User string `json:"user"`
	Msg  string `json:"msg"`
}
