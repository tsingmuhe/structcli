package structcli_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/tsingmuhe/structcli"
)

type MainCommand struct {
	Hello1 string  `short:"-h" long:"--hello" placeholder:"" description:"Hello"  `
	Hello2 *string `short:"-h" long:"--hello" placeholder:"" description:"Hello"  `
	Hello3 bool    `short:"-h" long:"--hello,negatable" description:"Hello"`
	Hello4 *bool   `short:"-h" long:"--hello,negatable" description:"Hello"`

	World1 string   `description:"Hello" placeholder:""`
	World2 *string  `description:"Hello" placeholder:""`
	World3 []string `description:"Hello" placeholder:""`

	SubCommand *SubCommand `command:"sub" description:"sub command"`
}

type SubCommand struct {
	Hello1 string  `short:"-h" long:"--hello" description:"Hello" placeholder:""`
	Hello2 *string `short:"-h" long:"--hello" description:"Hello" placeholder:""`
	Hello3 bool    `short:"-h" long:"--hello,negatable" description:"Hello"`
	Hello4 *bool   `short:"-h" long:"--hello,negatable" description:"Hello"`

	World1 string   `description:"Hello" placeholder:""`
	World2 *string  `description:"Hello" placeholder:""`
	World3 []string `description:"Hello" placeholder:""`
}

func TestCommandLine_Create(t *testing.T) {
	cli := structcli.Create("cmd", "a test cmd", "1.0.0", new(MainCommand))
	cmd, err := cli.Parse(os.Args[1:])
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(reflect.ValueOf(cmd).Type())
}
