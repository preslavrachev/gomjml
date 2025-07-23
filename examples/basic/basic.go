package main

import (
	"fmt"
	"log"

	"github.com/preslavrachev/gomjml/mjml"
	"github.com/preslavrachev/gomjml/parser"
)

func main() {
	mjmlContent := `<mjml>
        <mj-head>
          <mj-title>My Newsletter</mj-title>
        </mj-head>
        <mj-body>
          <mj-section>
            <mj-column>
              <mj-text>Hello World!</mj-text>
              <mj-button href="https://example.com">Click Me</mj-button>
            </mj-column>
          </mj-section>
        </mj-body>
      </mjml>`

	// Method 1: Direct rendering (recommended)
	html, err := mjml.Render(mjmlContent)
	if err != nil {
		log.Fatal("Render error:", err)
	}
	fmt.Println(html)

	// Method 2: Step-by-step processing
	ast, err := parser.ParseMJML(mjmlContent)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	component, err := mjml.CreateComponent(ast)
	if err != nil {
		log.Fatal("Component creation error:", err)
	}

	html, err = component.Render()
	if err != nil {
		log.Fatal("Render error:", err)
	}
	fmt.Println(html)
}
