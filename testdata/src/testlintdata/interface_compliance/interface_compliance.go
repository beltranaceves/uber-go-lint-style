package interface_compliance

import (
	"fmt"
	"net/http"
)

// BAD: exported type implements fmt.Stringer but no assertion
type BadStringer struct{} // want "exported type 'BadStringer' implements Stringer"

func (BadStringer) String() string {
    return fmt.Sprintf("bad")
}

// GOOD: compile-time assertion present
type GoodHandler struct{}

func (GoodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

var _ http.Handler = (*GoodHandler)(nil)
