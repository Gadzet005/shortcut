package multipartutils

import (
	"io"
	"mime/multipart"
	"net/textproto"

	"github.com/Gadzet005/shortcut/pkg/errors"
)

func WriteMultipartData(w io.Writer, data map[string][]byte) (string, error) {
	mw := multipart.NewWriter(w)
	defer mw.Close()

	for name, value := range data {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="`+name+`"`)

		part, err := mw.CreatePart(h)
		if err != nil {
			return "", errors.Wrapf(err, "create part for %s", name)
		}

		if _, err := part.Write(value); err != nil {
			return "", errors.Wrapf(err, "write part %s", name)
		}
	}

	return mw.FormDataContentType(), nil
}
