package example

type GetExampleRequest struct {
	Name string `json:"name"`
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
