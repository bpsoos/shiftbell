package profileselection

import (
	"net/http"
	"strconv"
)

const CookieName = "shiftbell_profile"

func SelectedId(request *http.Request) int {
	cookie, err := request.Cookie(CookieName)
	if err != nil {
		return 0
	}
	id, _ := strconv.Atoi(cookie.Value)
	return id
}
