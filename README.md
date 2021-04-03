# Colorizer

## Example Usage

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/airfork/colorizer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Example usage %s search terms\n", os.Args[0])
		os.Exit(0)
	}

	color, _ := colorizer.Colorize(strings.Join(os.Args[1:], " "))
	fmt.Println(color)
}
```

Visualize the color or convert to RGB at a site like [this](https://www.webfx.com/web-design/hex-to-rgb/)