package request

import "fmt"

type TemplateError struct {
	err  error
	file string
}

func (e TemplateError) Error() string {
	return fmt.Sprintf("failed to process template file %q: %v",
		e.file,
		e.err,
	)
}

type InvalidRequestFileError struct {
	msg string
}

func (e InvalidRequestFileError) Error() string {
	return e.msg
}

type InvalidRequestLineError struct {
	file string
	line string
}

func (e InvalidRequestLineError) Error() string {
	return fmt.Sprintf("invalid request line in file %q: expected %q, got %q",
		e.file,
		"URL, METHOD URL or METHOD URL PROTO",
		e.line,
	)
}

type InvalidURLError struct {
	err  error
	file string
	url  string
}

func (e InvalidURLError) Error() string {
	return fmt.Sprintf("failed to parse URL %q in file %q: %v",
		e.url,
		e.file,
		e.err,
	)
}

type InvalidHeaderError struct {
	err  error
	file string
}

func (e InvalidHeaderError) Error() string {
	return fmt.Sprintf("invalid header in file %q: %v",
		e.file,
		e.err,
	)
}

type BodyError struct {
	err  error
	file string
}

func (e BodyError) Error() string {
	return fmt.Sprintf("failed to parse body in file %q: %v",
		e.file,
		e.err,
	)
}
