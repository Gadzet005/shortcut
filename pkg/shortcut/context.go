package shortcut

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Context struct {
	base   *gin.Context
	items  map[string][]byte
	logger *zap.Logger
}

func newContext(c *gin.Context, logger *zap.Logger) *Context {
	return &Context{
		base:   c,
		items:  make(map[string][]byte),
		logger: logger,
	}
}

func (c *Context) parseRequest() error {
	if err := c.base.Request.ParseForm(); err != nil {
		return NewErrorWithCause(400, "failed to parse form data", err)
	}

	for key, values := range c.base.Request.PostForm {
		if len(values) != 0 {
			c.items[key] = []byte(values[0])
		}
	}

	return nil
}

func (c *Context) Logger() *zap.Logger {
	return c.logger
}

func (c *Context) GetItem(id string) ([]byte, bool) {
	data, ok := c.items[id]
	return data, ok
}

func (c *Context) GetItemJSON(id string, v any) error {
	data, ok := c.GetItem(id)
	if !ok {
		return ErrItemNotFound
	}

	if err := json.Unmarshal(data, v); err != nil {
		return NewErrorWithCause(400, "failed to unmarshal json", err)
	}

	return nil
}

func (c *Context) ListItems() []string {
	items := make([]string, 0, len(c.items))
	for id := range c.items {
		items = append(items, id)
	}
	return items
}

func (c *Context) HasItem(id string) bool {
	_, ok := c.items[id]
	return ok
}
