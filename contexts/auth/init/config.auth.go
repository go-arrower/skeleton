package init

type PWConfirmation struct {
	Active  bool
	Timeout int // how long the token is valid
}

type Config struct {
	InsecureAllowAnyPWStrength bool
	RegisterAllowed            bool //enabled / disabled
	PwHashCost                 int

	RegisterAdminRoutes bool

	LoginThrottle int // time in sec until a new login attempt can be made

	PWConfirmation PWConfirmation
	UserProvider   any // future music

	Mailer any // smtp <> local ect.
}
