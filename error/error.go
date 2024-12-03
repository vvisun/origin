package cherryError

import (
	"errors"
	"fmt"
)

func Error(text string) error {
	return errors.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func Wrap(err error, text string) error {
	return Errorf("err:%v, text:%s", err, text)
}

func Wrapf(err error, format string, a ...interface{}) error {
	text := fmt.Sprintf(format, a...)
	return Wrap(err, text)
}
