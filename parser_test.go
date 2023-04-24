package request

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParse(t *testing.T, rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	require.NoError(t, err)

	return u
}

func TestParser_Parse(t *testing.T) { //nolint:funlen
	p := Parser{Path: "testdata/"}
	tplData := struct {
		Host  string
		Key   string
		Value int
	}{
		Host:  "httpbin.org",
		Key:   "key",
		Value: 42,
	}
	tests := []struct {
		name        string
		fileName    string
		wantHeaders http.Header
		wantMethod  string
		wantProto   string
		wantURL     *url.URL
		wantBody    string
	}{
		{
			name:        "url",
			fileName:    "url.http",
			wantHeaders: http.Header{},
			wantMethod:  http.MethodGet,
			wantURL:     mustParse(t, "https://httpbin.org/get"),
		},
		{
			name:        "get",
			fileName:    "get.http",
			wantHeaders: http.Header{},
			wantMethod:  http.MethodGet,
			wantURL:     mustParse(t, "https://httpbin.org/get"),
		},
		{
			name:        "proto",
			fileName:    "get_proto.http",
			wantHeaders: http.Header{},
			wantMethod:  http.MethodGet,
			wantProto:   "HTTP/1.1",
			wantURL:     mustParse(t, "https://httpbin.org/get"),
		},
		{
			name:     "headers",
			fileName: "headers.http",
			wantHeaders: http.Header{
				"Accept":          {"application/json"},
				"Accept-Encoding": {"gzip, deflate, compress, br, *"},
			},
			wantMethod: http.MethodGet,
			wantURL:    mustParse(t, "https://httpbin.org/get"),
		},
		{
			name:     "post",
			fileName: "post.http",
			wantHeaders: map[string][]string{
				"Content-Type": {"application/json"},
			},
			wantMethod: http.MethodPost,
			wantURL:    mustParse(t, "https://httpbin.org/post"),
			wantBody: strings.Join(
				[]string{
					"{",
					`  "key": "value"`,
					"}",
					"",
				},
				"\r\n",
			),
		},
		{
			name:     "post from file",
			fileName: "post_from_file.http",
			wantHeaders: map[string][]string{
				"Content-Type": {"application/json"},
			},
			wantMethod: http.MethodPost,
			wantURL:    mustParse(t, "https://httpbin.org/post"),
			wantBody: strings.Join(
				[]string{
					"{",
					`  "key": 42`,
					"}",
					"\r\n",
				},
				"\n",
			),
		},
		{
			name:     "post template",
			fileName: "post_template.http",
			wantHeaders: map[string][]string{
				"Content-Type": {"application/json"},
			},
			wantMethod: http.MethodPost,
			wantURL:    mustParse(t, "https://httpbin.org/post"),
			wantBody: strings.Join(
				[]string{
					"{",
					`  "key": 42`,
					"}",
					"",
				},
				"\r\n",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.ParseSingle(tt.fileName, tplData)

			require.NoError(t, err)

			assert.Equal(t, tt.wantHeaders, got.Header)
			assert.Equal(t, tt.wantMethod, got.Method)
			assert.Equal(t, tt.wantProto, got.Proto)
			assert.Equal(t, tt.wantURL, got.URL)

			if tt.wantBody == "" {
				assert.Nil(t, got.Body)
			} else {
				body, err := io.ReadAll(got.Body)
				require.NoError(t, err)

				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestParser_Parse_Error(t *testing.T) {
	p := Parser{Path: "testdata/"}
	tplData := struct {
		Host  string
		Key   string
		Value int
	}{
		Host:  "httpbin.org",
		Key:   "key",
		Value: 42,
	}
	tests := []struct {
		name     string
		fileName string
		wantErr  string
	}{
		{
			name:     "empty",
			fileName: "empty.http",
			wantErr:  `no request in file "testdata/empty.http"`,
		},
		{
			name:     "invalid request line",
			fileName: "invalid_request_line.http",
			wantErr: `invalid request line in file "testdata/invalid_request_line.http": ` +
				`expected "URL, METHOD URL or METHOD URL PROTO", ` +
				`got "GET https://httpbin.org/get HTTP/1.1 EXTRA"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.fileName, tplData)

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestParser_Parse_Multiple(t *testing.T) { //nolint:funlen
	p := Parser{Path: "testdata/"}
	tplData := struct {
		Host  string
		Key   string
		Value int
	}{
		Host:  "httpbin.org",
		Key:   "key",
		Value: 42,
	}

	type want struct {
		headers http.Header
		method  string
		proto   string
		url     *url.URL
		body    string
	}

	wants := []want{
		{
			headers: http.Header{},
			method:  http.MethodGet,
			url:     mustParse(t, "https://httpbin.org/get"),
		},
		{
			headers: http.Header{},
			method:  http.MethodGet,
			url:     mustParse(t, "https://httpbin.org/get"),
		},
		{
			headers: http.Header{},
			method:  http.MethodGet,
			proto:   "HTTP/1.1",
			url:     mustParse(t, "https://httpbin.org/get"),
		},
		{
			headers: http.Header{
				"Accept":          {"application/json"},
				"Accept-Encoding": {"gzip, deflate, compress, br, *"},
			},
			method: http.MethodGet,
			url:    mustParse(t, "https://httpbin.org/get"),
		},
		{
			headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			method: http.MethodPost,
			url:    mustParse(t, "https://httpbin.org/post"),
			body: strings.Join(
				[]string{
					"{",
					`  "key": "value"`,
					"}",
					"",
					"",
				},
				"\r\n",
			),
		},
		{
			headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			method: http.MethodPost,
			url:    mustParse(t, "https://httpbin.org/post"),
			body: strings.Join(
				[]string{
					"{",
					`  "key": 42`,
					"}",
					"\r\n\r\n",
				},
				"\n",
			),
		},
		{
			headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			method: http.MethodPost,
			url:    mustParse(t, "https://httpbin.org/post"),
			body: strings.Join(
				[]string{
					"{",
					`  "key": 42`,
					"}",
					"",
				},
				"\r\n",
			),
		},
	}

	reqs, err := p.Parse("requests.http", tplData)
	require.NoError(t, err)
	require.Len(t, reqs, len(wants))

	for i, got := range reqs {
		w := wants[i]

		assert.Equal(t, w.headers, got.Header)
		assert.Equal(t, w.method, got.Method)
		assert.Equal(t, w.proto, got.Proto)
		assert.Equal(t, w.url, got.URL)

		if w.body == "" {
			assert.Nil(t, got.Body)
		} else {
			body, err := io.ReadAll(got.Body)
			require.NoError(t, err)

			assert.Equal(t, w.body, string(body))
		}
	}
}
