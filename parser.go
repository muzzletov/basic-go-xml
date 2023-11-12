package main

type Tag struct {
	Name       string
	Attributes map[string]string
	Body       string
	Children   []*Tag
	Namespace  string
}

type Parser struct {
	tags         []*Tag
	body         string
	root         *Tag
	current      *Tag
	namespaceTag *Tag
	namespaces   map[string]string
	length       int
}

func NewParser(body string) *Parser {
	parser := &Parser{
		body:       body,
		namespaces: make(map[string]string),
		length:     len(body),
		tags:       make([]*Tag, 0),
	}
	index := parser.consumeNamespaceTag(parser.skipWhitespace(0))
	parser.consumeTag(index)

	return parser
}

func (p *Parser) getCurrent() *Tag {
	return p.current
}

func (p *Parser) getNamespaceTag() *Tag {
	return p.namespaceTag
}

func (p *Parser) getTags() []*Tag {
	return p.tags
}

func (p *Parser) consumeNamespaceTag(index int) int {
	currentIndex := p.skipWhitespace(index)
	if p.body[currentIndex] != '<' {
		return -1
	}

	currentIndex++

	if p.body[currentIndex] != '?' {
		return -1
	}

	currentIndex++

	currentIndex = p.parseTagName(currentIndex)

	if currentIndex == -1 {
		return -1
	}

	isNamespaceTag := p.body[currentIndex] == '?' && p.body[currentIndex+1] == '>'

	if isNamespaceTag {
		p.namespaceTag = p.current
	}

	return currentIndex + 2
}

func (p *Parser) isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func (p *Parser) skipWhitespace(index int) int {
	if index == -1 {
		return index
	}

	for index < p.length && p.isWhitespace(rune(p.body[index])) {
		index++
	}

	return index
}

func (p *Parser) tagEnd(index int) int {
	if p.body[index] != '<' ||
		p.body[index+1] != '/' {
		return -1
	}

	length := len(p.current.Name) + index + 2

	for i := 0; i+index+2 < length; i++ {
		if p.body[index+i+2] != p.current.Name[i] {
			return -1
		}
	}

	if p.body[length] != '>' {
		return -1
	}

	return length + 1
}

func (p *Parser) consumeTag(index int) int {
	currentIndex := p.skipWhitespace(index)

	if currentIndex == -1 ||
		currentIndex >= p.length ||
		p.body[currentIndex] != '<' {
		return -1
	}

	currentIndex = p.parseTagName(currentIndex + 1)

	index = -1

	p.tags = append(p.tags, p.current)
	isEndOfTag := p.body[currentIndex] == '/' && p.body[currentIndex+1] == '>'

	if isEndOfTag {
		index += 2
	} else if p.body[currentIndex] == '>' {
		currentIndex = p.tagEnd(p.parseBody(currentIndex + 1))
		index = p.skipWhitespace(currentIndex)
	}

	current := p.tags[len(p.tags)-1]

	if len(p.tags) > 0 {
		p.current = p.tags[len(p.tags)-1]
		p.current.Children = append(p.current.Children, current)
	}
	return index
}

func (p *Parser) parseTagName(index int) int {
	currentIndex := index
	if p.ffLetter(index) == -1 {
		return -1
	}

	index = p.skipValidTag(currentIndex)
	current := &Tag{}

	if p.body[index] == ':' {
		current.Namespace = p.body[currentIndex:index]
		currentIndex = index + 1
		index = p.skipValidTag(index + 1)
	}

	current.Name = p.body[currentIndex:index]
	p.current = current
	current.Attributes = make(map[string]string)
	currentIndex = p.skipWhitespace(index)

	if currentIndex != index {
		currentIndex = p.parseAttributes(currentIndex)
	}

	return currentIndex
}

func (p *Parser) parseBody(index int) int {
	index = p.skipWhitespace(index)
	currentIndex := index
	current := p.current

	if p.body[index] != '<' {
		for p.body[index] != '<' {
			index++
		}

		current.Body = p.body[currentIndex:index]
		currentIndex = index
	} else {
		for index = p.consumeTag(index); index > -1; index = p.consumeTag(index) {
			index = p.skipWhitespace(index)
			currentIndex = index
			if p.body[index] == '<' &&
				p.body[index+1] == '/' {
				break
			}
		}
	}

	p.current = current

	return currentIndex
}

func (p *Parser) skipFF(index int) int {
	for p.ffLetter(index+1) != -1 {
		index++
	}

	return index + 1
}

func (p *Parser) parseAttributes(index int) int {
	currentIndex := index

	for currentIndex != -1 {
		var namespace *string = nil
		currentIndex = p.skipFF(currentIndex)
		name := p.body[index:currentIndex]

		if name == "xmlns" &&
			p.body[currentIndex] == ':' {
			index = currentIndex + 1
			currentIndex = p.skipFF(currentIndex + 1)
			temp := p.body[index:currentIndex]
			namespace = &temp
		}

		if p.body[currentIndex] != '=' &&
			p.body[currentIndex+1] != '"' {
			return -1
		}

		currentIndex = currentIndex + 2
		index = currentIndex

		for p.body[currentIndex] != '"' {
			currentIndex++
		}

		if namespace != nil {
			p.namespaces[*namespace] = p.body[index:currentIndex]
		} else {
			p.current.Attributes[name] = p.body[index:currentIndex]
		}

		currentIndex = p.skipWhitespace(currentIndex + 1)
		index = currentIndex

		if p.body[currentIndex] == '?' ||
			p.body[currentIndex] == '>' ||
			p.body[currentIndex] == '/' && p.body[currentIndex+1] == '>' {
			return index
		}
	}

	return currentIndex
}

func (p *Parser) ffLetter(index int) int {
	if index < p.length && isAlpha(rune(p.body[index])) {
		return index + 1
	}

	return -1
}

func isAlpha(r rune) bool {
	return ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}

func (p *Parser) skipValidTag(index int) int {
	for index < p.length && isValidTagChar(rune(p.body[index])) {
		index++
	}

	return index
}

func isValidTagChar(r rune) bool {
	return ('0' <= r && r <= '9') || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}
