# Selfor
Just Myself Context for Marcaron.

## Description
`Selfor` provides a [web.go][] compitable layer for reusing the code written with
hoisie's `web.go` framework. Here compitable means we can use `Selfor` the same 
way as in hoisie's `web.go` but not the others.

## Usage

~~~ go
package main

import (
	"github.com/insionng/selfor"
	"log"
	"os"
)

func main() {
	l, e := os.Create("./selfor.Classico.log")
	if e != nil {
		log.Fatal(e.Error())
	}
	defer l.Close()

	m := selfor.Classico(l)
	m.Get("/", func(self *selfor.Context) {
		self.WriteString("Hello , Hello , Hello , World!")
	})

	m.Run()
}

~~~

## Authors
* [Insion Ng](http://github.com/insionng)
* [Jeremy Saenz](http://github.com/codegangsta)
* [Archs Sun](http://github.com/Archs)
* [hoisie](https://github.com/hoisie)

## Links
* [Selfor Marcaron](http://github.com/insionng/selfor)
* [Self Martini](http://github.com/insionng/self)
