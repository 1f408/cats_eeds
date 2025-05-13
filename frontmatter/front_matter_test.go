package frontmatter_test

import (
	"fmt"
	"strings"

	"github.com/1f408/cats_eeds/frontmatter"
)

func ExampleMinus() {
	r := strings.NewReader("---\nhead1\nhead2\n---\nbody\n")

	fm := frontmatter.New([]string{"---", "+++"})

	wd, head, body, err := fm.FindAndSplit(r)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(wd == "---")
	fmt.Println(string(head) == "head1\nhead2\n")
	fmt.Println(string(body) == "body\n")
	// Output:
	// true
	// true
	// true
}

func ExamplePlus() {
	r := strings.NewReader("+++\r\nhead1\r\nhead2\r\n+++\r\nbody\r\n")

	fm := frontmatter.New([]string{"---", "+++"})

	wd, head, body, err := fm.FindAndSplit(r)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(wd == "+++")
	fmt.Println(string(head) == "head1\r\nhead2\r\n")
	fmt.Println(string(body) == "body\r\n")
	// Output:
	// true
	// true
	// true
}

func ExampleNoBody() {
	r := strings.NewReader("---\nhead1\nhead2\n---")

	fm := frontmatter.New([]string{"---", "+++"})

	wd, head, body, err := fm.FindAndSplit(r)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(wd == "---")
	fmt.Println(string(head) == "head1\nhead2\n")
	fmt.Println(string(body) == "")
	// Output:
	// true
	// true
	// true
}

func ExampleNoHead() {
	r := strings.NewReader("body\nbody\nbody\nbody\nbody\n")

	fm := frontmatter.New([]string{"---", "+++"})

	wd, _, _, err := fm.FindAndSplit(r)
	fmt.Println(err != nil)
	fmt.Println(wd == "")
	// Output:
	// true
	// true
}
