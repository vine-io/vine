package cli

import (
	"os"

	"github.com/vine-io/cli"
)

func quit(c *cli.Context, args []string) ([]byte, error) {
	os.Exit(0)
	return nil, nil
}

//func help(c *cli.Context, args []string) ([]byte, error) {
//	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
//
//	fmt.Fprintln(os.Stdout, "Commands:")
//
//	var keys []string
//	for k := range commands {
//		keys = append(keys, k)
//	}
//	sort.Strings(keys)
//
//	for _, k := range keys {
//		cmd := commands[k]
//		fmt.Fprintln(w, "\t", cmd.name, "\t\t", cmd.usage)
//	}
//
//	_ = w.Flush()
//	return nil, nil
//}
