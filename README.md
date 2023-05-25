# toy-dns

toy-dns is a toy dns resolver written to help familiarize me with Go and with DNS resolution.

It is used as follows:
```go
package main

import "fmt"

func main() {
	domain := "shopify.com"
	ip := resolve(domain, TYPE_A)
	fmt.Printf("IP address for '%s': %s\n", domain, ip)
}
```