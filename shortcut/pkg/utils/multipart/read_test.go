package multipartutils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadMultipartData(t *testing.T) {
	tests := []struct {
		name         string
		setupRequest func() (http.Header, io.Reader)
		expectedData map[string][]byte
		expectError  bool
	}{
		{
			name: "successful_read_single_field",
			setupRequest: func() (http.Header, io.Reader) {
				return createMultipartRequest(t, map[string]string{
					"field1": "value1",
				})
			},
			expectedData: map[string][]byte{
				"field1": []byte("value1"),
			},
		},
		{
			name: "successful_read_multiple_fields",
			setupRequest: func() (http.Header, io.Reader) {
				return createMultipartRequest(t, map[string]string{
					"field1": "value1",
					"field2": "value2",
					"field3": "value3",
				})
			},
			expectedData: map[string][]byte{
				"field1": []byte("value1"),
				"field2": []byte("value2"),
				"field3": []byte("value3"),
			},
		},
		{
			name: "empty_multipart",
			setupRequest: func() (http.Header, io.Reader) {
				return createMultipartRequest(t, map[string]string{})
			},
			expectedData: map[string][]byte{},
		},
		{
			name: "error_missing_content_type",
			setupRequest: func() (http.Header, io.Reader) {
				return createRequestWithContentType("")
			},
			expectedData: nil,
			expectError:  true,
		},
		{
			name: "error_not_multipart_content_type",
			setupRequest: func() (http.Header, io.Reader) {
				return createRequestWithContentType("application/json")
			},
			expectedData: nil,
			expectError:  true,
		},
		{
			name: "error_missing_boundary",
			setupRequest: func() (http.Header, io.Reader) {
				return createRequestWithContentType("multipart/form-data")
			},
			expectedData: nil,
			expectError:  true,
		},
		{
			name: "fields_with_empty_values",
			setupRequest: func() (http.Header, io.Reader) {
				return createMultipartRequest(t, map[string]string{
					"empty1": "",
					"empty2": "",
				})
			},
			expectedData: map[string][]byte{
				"empty1": []byte(""),
				"empty2": []byte(""),
			},
		},
		{
			name: "large_data",
			setupRequest: func() (http.Header, io.Reader) {
				largeData := string(bytes.Repeat([]byte("A"), 10000))
				return createMultipartRequest(t, map[string]string{
					"large": largeData,
				})
			},
			expectedData: map[string][]byte{
				"large": bytes.Repeat([]byte("A"), 10000),
			},
		},
		{
			name: "special_characters_in_values",
			setupRequest: func() (http.Header, io.Reader) {
				return createMultipartRequest(t, map[string]string{
					"special": "Hello\nWorld\t\r\n",
					"unicode": "Привет мир 🌍",
				})
			},
			expectedData: map[string][]byte{
				"special": []byte("Hello\nWorld\t\r\n"),
				"unicode": []byte("Привет мир 🌍"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, body := tt.setupRequest()

			result, err := ReadMultipartData(header, body)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, len(tt.expectedData), len(result))

			for key, expectedValue := range tt.expectedData {
				actualValue, exists := result[key]
				require.True(t, exists)
				require.Equal(t, expectedValue, actualValue)
			}
		})
	}
}

func createMultipartRequest(t *testing.T, fields map[string]string) (http.Header, io.Reader) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for name, value := range fields {
		err := writer.WriteField(name, value)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	header := http.Header{}
	header.Set("Content-Type", writer.FormDataContentType())

	return header, body
}

func createRequestWithContentType(contentType string) (http.Header, io.Reader) {
	header := http.Header{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	body := &bytes.Buffer{}
	return header, body
}
