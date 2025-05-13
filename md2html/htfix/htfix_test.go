package htfix_test

import (
	"fmt"

	"github.com/1f408/cats_eeds/md2html/htfix"
)

func ExampleString() {
	src := []string{
		"<span>hoge",
		"<span>hoge</span>",
		"<html><body><span>hoge</span></body></html>",
	}
	for _, s := range src {
		fmt.Println(htfix.FixHTMLString(s))
	}
	// Output:
	// <span>hoge</span>
	// <span>hoge</span>
	// <span>hoge</span>
}

func ExampleBytes() {
	src := [][]byte{
		[]byte("<span>hoge"),
		[]byte("<span>hoge</span>"),
		[]byte("<html><body><span>hoge</span></body></html>"),
	}
	for _, s := range src {
		fmt.Println(string(htfix.FixHTML(s)))
	}
	// Output:
	// <span>hoge</span>
	// <span>hoge</span>
	// <span>hoge</span>
}

func ExampleMultiString() {
	svg := `<svg class="icon" fill="none" stroke-width="1.5" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
  <path d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" stroke-linecap="round" stroke-linejoin="round"></path>
`
	fmt.Print(htfix.FixHTMLString(svg))
	// Output:
	// <svg class="icon" fill="none" stroke-width="1.5" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
	//   <path d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" stroke-linecap="round" stroke-linejoin="round"></path>
	// </svg>
}

func ExampleMultiBytes() {
	svg := []byte(`<svg class="icon" fill="none" stroke-width="1.5" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
  <path d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" stroke-linecap="round" stroke-linejoin="round"></path>
`)
	fmt.Print(string(htfix.FixHTML(svg)))
	// Output:
	// <svg class="icon" fill="none" stroke-width="1.5" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
	//   <path d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" stroke-linecap="round" stroke-linejoin="round"></path>
	// </svg>
}
