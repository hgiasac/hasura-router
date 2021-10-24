package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapError(t *testing.T) {

	err := Error{
		Code:    "foo",
		Message: "bar",
	}

	var unwrappedError Error

	if !errors.As(err, &unwrappedError) {
		t.Fatal("invalid action error")
	}

	assert.Equal(t, err.Code, unwrappedError.Code)
	assert.Equal(t, err.Message, err.Message)

}
