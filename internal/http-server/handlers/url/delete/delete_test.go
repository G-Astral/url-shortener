package delete_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func TestDeleteHandler(t *testing.T) {
	const alias = "test-alias"

	tests := []struct {
		name         string
		aliasInURL   string
		mockSetup    func(m *mocks.URLDeletter)
		expectedCode int
		expectedBody string
	}{
		{
			name:       "success",
			aliasInURL: alias,
			mockSetup: func(m *mocks.URLDeletter) {
				m.On("DeleteURL", alias).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: fmt.Sprintf(`"url by alias: %s deleted"`, alias),
		},
		{
			name:       "alias is empty",
			aliasInURL: "",
			mockSetup:  func(m *mocks.URLDeletter) {},
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:       "url not found",
			aliasInURL: alias,
			mockSetup: func(m *mocks.URLDeletter) {
				m.On("DeleteURL", alias).Return(storage.ErrURLNotFound).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"error":"not found", "status":"Error"}`,

		},
		{
			name:       "internal error",
			aliasInURL: alias,
			mockSetup: func(m *mocks.URLDeletter) {
				m.On("DeleteURL", alias).Return(errors.New("some internal error")).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"error":"internal error", "status":"Error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
			mock := mocks.NewURLDeletter(t)
			tt.mockSetup(mock)

			r := chi.NewRouter()
			r.Delete("/{alias}", delete.New(slogdiscard.NewDiscardLogger(), mock))

			req := httptest.NewRequest(http.MethodDelete, "/"+tt.aliasInURL, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, res.StatusCode)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}
