package custom

// JSONResponse is template for all the api responses
type JSONResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// RequestError implements error interface for request related errors
type RequestError struct {
	Status  int // https status code
	Message string
}

func (err RequestError) Error() string {
	return err.Message
}

// PageInfo defines pagination info for list of data
type PageInfo struct {
	NextPage bool   `json:"nextPage"`
	Cursor   string `json:"cursor"`
}

// PaginatedData is data form JSONResponse for list of data
type PaginatedData struct {
	PageInfo PageInfo `json:"pageInfo"`
	Items    any      `json:"items"`
}