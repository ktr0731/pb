package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/golang/protobuf/proto"  //nolint:staticcheck
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic/msgregistry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var importPaths, importFiles []string

var (
	rootCmd = &cobra.Command{
		Use:               "pb",
		Short:             "Protocol Buffers utility",
		PersistentPreRunE: initMessageRegistry,
	}

	lsCmd = &cobra.Command{
		Use:   "ls <descriptor type>",
		Short: "list loaded top-level descriptors",
		RunE:  list,
	}
)

func main() {
	rootCmd.AddCommand(lsCmd, newDecodeCommand())
	rootCmd.PersistentFlags().StringSliceVarP(&importPaths, "proto_path", "I", nil, "import paths")
	rootCmd.PersistentFlags().StringSliceVarP(&importFiles, "proto_file", "F", nil, "import files")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "pb: %s\n", err)
		os.Exit(1)
	}
}

var (
	fileDescs []*desc.FileDescriptor
	msgReg    *msgregistry.MessageRegistry
)

func initMessageRegistry(_ *cobra.Command, args []string) error {
	parser := &protoparse.Parser{
		ImportPaths: importPaths,
	}
	fds, err := parser.ParseFiles(importFiles...)
	if err != nil {
		return errors.Wrap(err, "failed to parse proto files")
	}
	fileDescs = fds

	msgReg = msgregistry.NewMessageRegistryWithDefaults()
	for _, fd := range fds {
		msgReg.AddFile("", fd)
	}
	return nil
}

func list(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New(`specify one of "files", "messages" or "services"`)
	}

	m := map[string]struct{}{}
	for _, a := range args {
		m[a] = struct{}{}
	}

	for k := range m {
		switch strings.ToLower(k) {
		case "file", "files":
			for _, fd := range fileDescs {
				fmt.Println(fd.GetFullyQualifiedName())
			}
		case "msg", "message", "messages":
			for _, fd := range fileDescs {
				for _, mt := range fd.GetMessageTypes() {
					fmt.Println(mt.GetFullyQualifiedName())
				}
			}
		case "svc", "services":
			for _, fd := range fileDescs {
				for _, svc := range fd.GetServices() {
					fmt.Println(svc.GetFullyQualifiedName())
				}
			}
		}
	}

	return nil
}

func newDecodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode <message>",
		Short: "decode base64-encoded input as a JSON string",
		RunE:  decode,
	}
	cmd.Flags().String("in", "bin", `input type. "bin" or "base64".`)
	return cmd
}

func decode(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New(`specify fully-qualified message name`)
	}

	msg, err := msgReg.Resolve(args[0])
	if err != nil {
		return errors.Wrap(err, "failed to resolve message")
	}

	var r io.Reader = os.Stdin
	if t, _ := cmd.Flags().GetString("in"); t == "base64" {
		r = base64.NewDecoder(base64.StdEncoding, r)
	}

	in, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "failed to read base64-encoded input from stdin")
	}
	if err := proto.Unmarshal(in, msg); err != nil {
		return errors.Wrap(err, "failed to unmarshal message")
	}

	ms := &jsonpb.Marshaler{
		Indent: "  ",
	}
	if err := ms.Marshal(os.Stdout, msg); err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}
	fmt.Println()

	return nil
}
