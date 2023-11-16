package ssh

type Credential struct {
	Key      string
	Password string
}

func (c Credential) IsEmpty() bool {
	return c.Key == "" && c.Password == ""
}
