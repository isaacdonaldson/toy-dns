package main

import "fmt"

func main() {
	domain := "shopify.com"
	ip := resolve(domain, TYPE_A)
	fmt.Printf("IP address for '%s': %s\n", domain, ip)
}
