package multipartutils

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/Gadzet005/shortcut/pkg/errors"
)

func ReadMultipartData(header http.Header, body io.Reader) (map[string][]byte, error) {
	contentType := header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, errors.Wrap(err, "parse content-type")
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, errors.Errorf("not a multipart response: %s", mediaType)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, errors.Error("boundary not found")
	}

	reader := multipart.NewReader(body, boundary)

	result := make(map[string][]byte)
	for {
		part, err := reader.NextPart()
		switch {
		case errors.Is(err, io.EOF):
			return result, nil
		case err != nil:
			return nil, errors.Wrap(err, "read part")
		}

		body, err := io.ReadAll(part)
		if err != nil {
			return nil, errors.Wrap(err, "read part body")
		}

		result[part.FormName()] = body
	}
}
