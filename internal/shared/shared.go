package shared

type ConfigRepository struct {
	Name     string
	Ref      string
	Key      string
	Password string
}

func (c *ConfigRepository) IsEmpty() bool {
	return c.Key == "" && c.Password == ""
}
