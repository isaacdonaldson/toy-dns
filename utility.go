package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

func headerToBytes(header DnsHeader) []byte {
	return intsToHexBytes([]int{header.id, header.flags, header.num_questions, header.num_answers, header.num_authorities, header.num_additionals})
}

func questionToBytes(question DnsQuestion) []byte {
	return append([]byte(question.name), intsToHexBytes([]int{question.type_, question.class})...)
}

func decodeName(reader io.Reader) []byte {
	parts := []byte{}
	length_bytes := make([]byte, 1)

	for {
		// Initialize for each loop
		_, err := reader.Read(length_bytes)
		if err != nil {
			panic(err)
		}
		length := length_bytes[0]

		// break condition
		if length == 0 {
			break
		}

		// logic
		if length&0b11000000 == 0b11000000 {
			part := decodeCompressedName(length, reader)
			parts = append(parts, part...)
			parts = append(parts, '.')
			break
		} else {
			part := make([]byte, length)
			_, err := reader.Read(part)
			if err != nil {
				panic(err)
			}
			parts = append(parts, part...)
		}

		parts = append(parts, '.')
	}

	if len(parts) == 0 {
		return parts
	} else {
		return parts[:len(parts)-1]
	}
}

func decodeCompressedName(length byte, reader io.Reader) []byte {
	byteAddition := make([]byte, 1)
	_, err := reader.Read(byteAddition)
	if err != nil {
		panic(err)
	}

	masked := byte(length & 0b00111111)
	pointerBytes := []byte{masked, byteAddition[0]}
	pointer := int(binary.BigEndian.Uint16(pointerBytes))
	currentPos, err := reader.(io.Seeker).Seek(0, io.SeekCurrent)
	if err != nil {
		panic(err)
	}

	_, err = reader.(io.Seeker).Seek(int64(pointer), io.SeekStart)
	if err != nil {
		panic(err)
	}

	name := decodeName(reader)

	_, err = reader.(io.Seeker).Seek(currentPos, io.SeekStart)
	if err != nil {
		panic(err)
	}

	return name
}

func encodeDnsName(domain string) []byte {
	var accum bytes.Buffer
	for _, str := range strings.Split(domain, ".") {
		length := uint8(len(str))
		accum.WriteByte(length)
		for idx := 0; idx < len(str); idx++ {
			accum.WriteByte(str[idx])
		}
	}
	accum.WriteByte(uint8(0))

	return accum.Bytes()
}

func build_query(domain string, recordType int) []byte {
	name := encodeDnsName(domain)
	id := rand.New(rand.NewSource(1)).Intn(65535)
	header := DnsHeader{id: id, flags: 0, num_questions: 1} // Everything else defaults to 0
	question := DnsQuestion{name: name, type_: recordType, class: CLASS_IN}
	return append(headerToBytes(header), questionToBytes(question)...)
}

func intsToHexBytes(nums []int) []byte {
	buf := make([]byte, 2)
	var accum bytes.Buffer
	for _, n := range nums {
		n := uint16(n)
		binary.BigEndian.PutUint16(buf, n)
		accum.WriteByte(buf[0])
		accum.WriteByte(buf[1])
	}

	return accum.Bytes()
}

func sendQuery(ipAddr, domain string, recordType int) DnsPacket {
	req := build_query(domain, recordType)

	con, err := net.Dial("udp", fmt.Sprintf("%s:53", ipAddr))
	if err != nil {
		panic(err)
	}
	defer con.Close()

	_, err = con.Write(req)
	if err != nil {
		panic(err)
	}

	res := make([]byte, 1024)

	_, err = con.Read(res)
	if err != nil {
		panic(err)
	}

	return parseDnsPacket(res)
}

func ipToString(ip []byte) string {
	numbers := make([]string, 4)
	for idx, el := range ip {
		numbers[idx] = strconv.Itoa(int(el))
	}
	return strings.Join(numbers, ".")
}

func getAnswer(packet DnsPacket) []byte {
	var returnVal []byte
	for _, record := range packet.answers {
		if record.type_ == TYPE_A {
			returnVal = record.data
			break
		}
	}

	return returnVal
}

func getNameserverIp(packet DnsPacket) []byte {
	var returnVal []byte
	for _, record := range packet.additionals {
		if record.type_ == TYPE_A {
			returnVal = record.data
			break
		}
	}

	return returnVal
}

func getNameserver(packet DnsPacket) string {
	var returnVal string
	for _, record := range packet.authorities {
		if record.type_ == TYPE_NS {
			returnVal = string(record.data)
			break
		}
	}

	return returnVal
}

func resolve(domain string, recordType int) string {
	nameserver := "198.41.0.4"

	for {
		fmt.Printf("Querying %s for %s\n", nameserver, domain)
		res := sendQuery(nameserver, domain, recordType)

		if ip := getAnswer(res); ip != nil {
			return ipToString(ip)
		} else if nsIp := getNameserverIp(res); nsIp != nil {
			nameserver = ipToString(getNameserverIp(res))
		} else if nsDomain := getNameserver(res); nsDomain != "" {
			nameserver = resolve(nsDomain, TYPE_A)
		} else {
			panic("No answer or nameserver found")
		}
	}
}
