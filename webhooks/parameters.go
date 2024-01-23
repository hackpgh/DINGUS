package webhooks

type Parameter interface {
	Validate() error
	String() string
	ToJSON() ([]byte, error)
}
