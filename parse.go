package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

func parseDnsPacket(responseData []byte) DnsPacket {
	reader := bytes.NewReader(responseData)
	header := parseHeader(reader)
	var questions []DnsQuestion
	for i := 0; i < header.num_questions; i++ {
		questions = append(questions, parseQuestion(reader))
	}
	var answers []DnsRecord
	for i := 0; i < header.num_answers; i++ {
		answers = append(answers, parseRecord(reader))
	}
	var authorities []DnsRecord
	for i := 0; i < header.num_authorities; i++ {
		authorities = append(authorities, parseRecord(reader))
	}
	var additionals []DnsRecord
	for i := 0; i < header.num_additionals; i++ {
		additionals = append(additionals, parseRecord(reader))
	}

	return DnsPacket{
		header:      header,
		questions:   questions,
		answers:     answers,
		authorities: authorities,
		additionals: additionals,
	}
}

func parseHeader(reader io.Reader) DnsHeader {
	data := make([]byte, 12)
	_, err := reader.Read(data)
	if err != nil {
		panic(err)
	}

	return DnsHeader{
		id:              int(binary.BigEndian.Uint16(data[0:2])),
		flags:           int(binary.BigEndian.Uint16(data[2:4])),
		num_questions:   int(binary.BigEndian.Uint16(data[4:6])),
		num_answers:     int(binary.BigEndian.Uint16(data[6:8])),
		num_authorities: int(binary.BigEndian.Uint16(data[8:10])),
		num_additionals: int(binary.BigEndian.Uint16(data[10:12])),
	}
}

func parseQuestion(reader io.Reader) DnsQuestion {
	name := decodeName(reader)
	data := make([]byte, 4)
	_, err := reader.Read(data)
	if err != nil {
		panic(err)
	}

	return DnsQuestion{
		name:  name,
		type_: int(binary.BigEndian.Uint16(data[0:2])),
		class: int(binary.BigEndian.Uint16(data[2:4])),
	}
}

func parseRecord(reader io.Reader) DnsRecord {
	name := decodeName(reader)

	data := make([]byte, 10)
	_, err := reader.Read(data)
	if err != nil {
		panic(err)
	}

	type_ := int(binary.BigEndian.Uint16(data[0:2]))
	class := int(binary.BigEndian.Uint16(data[2:4]))
	ttl := int(binary.BigEndian.Uint32(data[4:8]))
	data_length := int(binary.BigEndian.Uint16(data[8:10]))

	var data_ []byte

	if type_ == TYPE_NS {
		data_ = decodeName(reader)
	} else if type_ == TYPE_A {
		data_ = make([]byte, data_length)
		_, err := reader.Read(data_)
		if err != nil {
			panic(err)
		}
	} else {
		data_ := make([]byte, data_length)
		_, err = reader.Read(data_)
		if err != nil {
			panic(err)
		}
	}

	return DnsRecord{
		name:  name,
		type_: type_,
		class: class,
		ttl:   ttl,
		data:  data_,
	}
}
