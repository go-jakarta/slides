import "net/http"

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
