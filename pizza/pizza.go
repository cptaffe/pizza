package pizza

type Addr struct {
	Street string `yaml:"street"`
	City   string `yaml:"city"`
	State  string `yaml:"state"`
	Zip    string `yaml:"zip"`
}

type Store interface {
	Addr() (Addr, error)
}
