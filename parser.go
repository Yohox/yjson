package main

import (
	"fmt"
	"io"
)

const (
	OB = '{'
	CB = '}'
	LB = '['
	RB = ']'
	DQ = '"'
	BLANK_SPACE = ' '
	HORIZONTAL_TAB = '\t'
	LINE_BREAK = '\n'
	CARRIAGE_RETURN = '\r'
	VALUE_SEPARATOR = ':'
	FALSE = "false"
	TRUE = "true"
	NULL = "null"
	DOT = ','
)

type Parser struct {
	buf []byte
	i int
	len int
}

const (
	JSON_NUMBER = iota
	JSON_STRING
	JSON_BOOLEAN
	JSON_OBJECT
	JSON_ARRAY
	JSON_NULL
)

type JsonValue struct {
	valueType int
	value interface{}
}

func (p *Parser) expect(b byte) error {
	peak, err := p.peak()
	if err != nil {
		return err
	}
	if peak != b {
		return fmt.Errorf("expect: %b, but get: %b", b, peak)
	}
	return nil
}

func (p *Parser) expectString(str string) error {
	if p.i + len(str) >= p.len {
		return io.EOF
	}
	s := p.buf[p.i: p.i+len(str)]
	if string(s) != str {
		return fmt.Errorf("expect: %s, but get: %s", str, s)
	}
	return nil
}

func (p *Parser) peak() (byte, error) {
	if p.i == p.len {
		return 0, io.EOF
	}

	return p.buf[p.i], nil
}

func (p *Parser) readAt(buf []byte, pos int) (int, error) {
	if pos + len(buf) >= p.len {
		return 0, io.EOF
	}

	n := copy(buf, p.buf[pos:pos+len(buf)])
	return n, nil
}

func (p *Parser) init(j *JsonValue) error {
	err := p.expect(LB)
	if err != nil {
		err = p.expect(LB)
		if err != nil {
			return fmt.Errorf("not found { or [")
		}
	}

	return p.handle(j)
}

func (p* Parser) readByte() (byte, error) {
	if p.i >= p.len {
		return 0, io.EOF
	}
	b := p.buf[p.i]
	p.i++
	return b, nil
}

func (p *Parser) absorbLack() error {
	b, err := p.peak()
	if err != nil {
		return err
	}

	for b == BLANK_SPACE || b == HORIZONTAL_TAB || b == LINE_BREAK  || b == CARRIAGE_RETURN {
		_, err := p.readByte()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) absorbByte(b byte) error {
	err := p.expect(b)
	if err != nil {
		return err
	}

	p.i++
	return nil
}

func (p *Parser) parseString(j *JsonValue) error {
	err := p.absorbByte(DQ)
	if err != nil {
		return err
	}

	str := make([]byte, 0)

	b, err := p.peak()
	if err != nil {
		return err
	}

	for true {
		b, err = p.peak()
		if err != nil {
			return err
		} else if b == DQ {
			p.i++
			break
		}

		b, err = p.readByte()
		if err != nil {
			return err
		}
		str = append(str, b)

	}


	j.valueType = JSON_STRING
	j.value = string(str)
	return nil
}

func (p *Parser) handle(j *JsonValue) error {
	err := p.absorbLack()
	if err != nil {
		return err
	}

	b, err := p.peak()
	switch b {
	case OB: // 左花括号
		err := p.parseObject(j)
		if err != nil {
			return err
		}
	case LB: // 左中括号
		err := p.parseArray(j)
		if err != nil {
			return err
		}
	case DQ: // 字符串
		err := p.parseString(j)
		if err != nil {
			return  err
		}
	case 'f':
		err := p.expectString(FALSE)
		if err != nil {
			return err
		}
		p.i += len(FALSE)
		j.valueType = JSON_BOOLEAN
		j.value = true
	case 'n':
		err := p.expectString(NULL)
		if err != nil {
			return err
		}
		p.i += len(NULL)
		j.valueType = JSON_NULL
	case 't':
		err := p.expectString(TRUE)
		if err != nil {
			return err
		}
		p.i += len(TRUE)
		j.valueType = JSON_BOOLEAN
		j.value = false
	default:
		return fmt.Errorf("not match")
	}

	return nil
}

func (p *Parser) parseObject(j *JsonValue) error {
	err := p.absorbByte(OB)
	if err != nil {
		return err
	}

	jsonObjectMap := make(map[string]*JsonValue)

	for true {

		err = p.absorbLack()
		if err != nil {
			return err
		}

		b, err := p.peak()
		if err != nil {
			return err
		}

		if b == CB {
			p.i++
			break
		}

		key := &JsonValue{}
		value := &JsonValue{}
		err = p.parseString(key)
		if err != nil {
			return err
		}

		err = p.absorbLack()
		if err != nil {
			return err
		}

		err = p.absorbByte(VALUE_SEPARATOR)


		if err != nil {
			return err
		}

		err = p.absorbLack()
		if err != nil {
			return err
		}

		err = p.handle(value)
		if err != nil {
			return err
		}

		jsonObjectMap[key.value.(string)] = value
	}

	j.valueType = JSON_OBJECT
	j.value = jsonObjectMap
	return nil
}

func (p *Parser) parseArray(j *JsonValue) error {
	arr := make([]interface{}, 0)
	err := p.absorbByte(LB)
	if err != nil {
		return err
	}


	for true {

		err = p.absorbLack()
		if err != nil {
			return err
		}

		b, err := p.peak()
		if err != nil {
			return err
		}

		if b == RB {
			p.i++
			break
		}
		value := &JsonValue{}
		err = p.handle(value)
		if err != nil {
			return err
		}

		err = p.absorbLack()
		if err != nil {
			return err
		}

		b, err = p.peak()
		if err != nil {
			return err
		}

		if b == DOT {
			p.i++
		}
	}

	j.valueType = JSON_ARRAY
	j.value = arr
	return nil
}


func Marshal(data []byte) (*JsonValue, error) {
	parser := &Parser{buf: data, len: len(data)}
	res := &JsonValue{}
	err := parser.init(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}