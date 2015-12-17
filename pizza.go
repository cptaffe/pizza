package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cptaffe/pizza/dominos"
	"github.com/cptaffe/pizza/pizza"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Addr pizza.Addr `yaml:"address"`
}

func main() {
	c := Config{}
	b, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(b, &c); err != nil {
		log.Fatal(err)
	}
	stores, err := dominos.Stores(&c.Addr)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range stores {
		a, err := s.Addr()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%#v\n", a)
	}
}
