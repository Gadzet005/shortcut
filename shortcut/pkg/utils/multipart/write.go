package multipartutils

import (
	"io"
	"mime/multipart"
	"net/textproto"

	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
)

func WriteMultipartData(w io.Writer, data map[string][]byte) (string, error) {
	mw := multipart.NewWriter(w)
	defer mw.Close()

	for name, value := range data {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="`+name+`"`)

		part, err := mw.CreatePart(h)
		if err != nil {
			return "", errorsutils.WrapFail(err, "create part for %s", name)
		}

		if _, err := part.Write(value); err != nil {
			return "", errorsutils.WrapFail(err, "write part %s", name)
		}
	}

	return mw.FormDataContentType(), nil
}
