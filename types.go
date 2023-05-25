package main

const TYPE_A = 1
const TYPE_NS = 2
const TYPE_TXT = 16
const CLASS_IN = 1

type DnsRecord struct {
	name  []byte
	type_ int
	class int
	ttl   int
	data  []byte
}

type DnsPacket struct {
	header      DnsHeader
	questions   []DnsQuestion
	answers     []DnsRecord
	authorities []DnsRecord
	additionals []DnsRecord
}

type DnsHeader struct {
	id              int
	flags           int
	num_questions   int
	num_answers     int
	num_authorities int
	num_additionals int
}

type DnsQuestion struct {
	name  []byte
	type_ int
	class int
}
