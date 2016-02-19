package controllers

type AuthController struct {
	AB backends.AuthBackend
}

func GetNewAuthenticationController(authBackend backends.AuthBackend) *AuthController {
	return &AuthController{AB: authBackend}
}

func (a *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	requestUser := new(authentication.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	responseStatus, token := authentication.Login(requestUser, a.AB)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseStatus)
	w.Write(token)
}

func (a *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestUser := new(authentication.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	w.Header().Set("Content-Type", "application/json")
	w.Write(authentication.RefreshToken(requestUser, a.AB))
}

