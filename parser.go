package request

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	requestLineURL            = 1
	requestLineMethodURL      = 2
	requestLineMethodURLProto = 3
)

const httpEOL = "\r\n"

type Parser struct {
	BaseURL string
	Path    string
}

func (p Parser) Parse(fileName string, data any) (*http.Request, error) {
	buf, err := p.exec(fileName, data)
	if err != nil {
		return nil, err
	}

	r := textproto.NewReader(bufio.NewReader(buf))

	req, err := p.parseRequest(fileName, r)
	if err != nil {
		return nil, err
	}

	err = p.parseHeaders(r, req)
	if errors.Is(err, io.EOF) {
		return req, nil
	}

	if err != nil {
		return nil, InvalidHeaderError{
			err:  err,
			file: p.file(fileName),
		}
	}

	if err = p.parseBody(fileName, r, req, data); err != nil {
		return nil, err
	}

	return req, nil
}

func (p Parser) exec(fileName string, data any) (*bytes.Buffer, error) {
	file := p.file(fileName)

	tpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, TemplateError{
			err:  err,
			file: file,
		}
	}

	result := &bytes.Buffer{}

	err = tpl.Execute(result, data)
	if err != nil {
		return nil, TemplateError{
			err:  err,
			file: file,
		}
	}

	return result, nil
}

func (p Parser) parseRequest(fileName string, r *textproto.Reader) (*http.Request, error) {
	for {
		line, err := r.ReadContinuedLine()
		if err != nil {
			return nil, InvalidRequestFileError{
				err:  err,
				file: p.file(fileName),
			}
		}

		if isComment(line) {
			continue
		}

		return p.parseRequestLine(fileName, line)
	}
}

func (p Parser) parseRequestLine( //nolint:nonamedreturns
	fileName string,
	line string,
) (req *http.Request, err error) {
	fields := strings.Fields(line)

	switch len(fields) {
	case requestLineURL:
		req = &http.Request{Method: http.MethodGet}
		req.URL, err = p.parseURL(fileName, fields[0])
	case requestLineMethodURL:
		req = &http.Request{Method: fields[0]}
		req.URL, err = p.parseURL(fileName, fields[1])
	case requestLineMethodURLProto:
		req = &http.Request{
			Method: fields[0],
			Proto:  fields[2],
		}
		req.URL, err = p.parseURL(fileName, fields[1])
	default:
		err = InvalidRequestLineError{
			file: p.file(fileName),
			line: line,
		}
	}

	return req, err
}

func (p Parser) parseURL(fileName string, s string) (*url.URL, error) {
	rawURL := p.BaseURL + s

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, InvalidURLError{
			err:  err,
			file: p.file(fileName),
			url:  rawURL,
		}
	}

	return u, nil
}

func (p Parser) parseBody(fileName string, r *textproto.Reader, req *http.Request, data any) error {
	sb := &strings.Builder{}

	for {
		line, err := r.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return BodyError{
				err:  err,
				file: p.file(fileName),
			}
		}

		if strings.HasPrefix(line, "<") {
			if err = p.appendFile(sb, strings.TrimSpace(line[1:]), data); err != nil {
				return err
			}

			continue
		}

		sb.WriteString(line)
		sb.WriteString(httpEOL)
	}

	if sb.Len() > 0 {
		req.Body = io.NopCloser(bytes.NewBufferString(sb.String()))
	}

	return nil
}

func (p Parser) appendFile(sb *strings.Builder, fileName string, data any) error {
	buf, err := p.exec(fileName, data)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(sb)
	if err != nil {
		return BodyError{
			err:  err,
			file: p.file(fileName),
		}
	}

	sb.WriteString(httpEOL)

	return nil
}

func (p Parser) file(fileName string) string {
	return filepath.Join(p.Path, fileName)
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}

func (p Parser) parseHeaders(r *textproto.Reader, req *http.Request) error {
	h, err := r.ReadMIMEHeader()

	req.Header = http.Header(h)

	return err //nolint:wrapcheck
}
