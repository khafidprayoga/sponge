package generate

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/zhufuyi/sponge/pkg/gofile"
	"github.com/zhufuyi/sponge/pkg/replacer"
	"github.com/zhufuyi/sponge/pkg/sql2code"
	"github.com/zhufuyi/sponge/pkg/sql2code/parser"

	"github.com/spf13/cobra"
)

// HandlerPbCommand generate handler and protobuf codes
func HandlerPbCommand() *cobra.Command {
	var (
		moduleName string // module name for go.mod
		serverName string // server name
		outPath    string // output directory
		dbTables   string // table names

		sqlArgs = sql2code.Args{
			Package:  "model",
			JSONTag:  true,
			GormType: true,
		}
	)

	cmd := &cobra.Command{
		Use:   "handler-pb",
		Short: "Generate handler and protobuf codes based on mysql table",
		Long: `generate handler and protobuf codes based on mysql table.

Examples:
  # generate handler and protobuf codes and embed 'gorm.model' struct.
  sponge web handler-pb --module-name=yourModuleName --server-name=yourServerName --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user

  # generate handler and protobuf codes with multiple table names.
  sponge web handler-pb --module-name=yourModuleName --server-name=yourServerName --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=t1,t2

  # generate handler and protobuf codes, structure fields correspond to the column names of the table.
  sponge web handler-pb --module-name=yourModuleName --server-name=yourServerName --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --embed=false

  # generate handler and protobuf codes and specify the server directory, Note: code generation will be canceled when the latest generated file already exists.
  sponge web handler-pb --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --out=./yourServerDir
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdName, srvName := getNamesFromOutDir(outPath)
			if mdName != "" {
				moduleName = mdName
			} else if moduleName == "" {
				return errors.New(`required flag(s) "module-name" not set, use "sponge web handler-pb -h" for help`)
			}
			if srvName != "" {
				serverName = srvName
			} else if serverName == "" {
				return errors.New(`required flag(s) "server-name" not set, use "sponge web handler-pb -h" for help`)
			}

			tableNames := strings.Split(dbTables, ",")
			for _, tableName := range tableNames {
				if tableName == "" {
					continue
				}

				sqlArgs.DBTable = tableName
				codes, err := sql2code.Generate(&sqlArgs)
				if err != nil {
					return err
				}

				outPath, err = runGenHandlerPbCommand(moduleName, serverName, codes, outPath)
				if err != nil {
					return err
				}
			}

			fmt.Printf(`
using help:
  1. move the folders "api" and "internal" to your project code folder.
  2. open a terminal and execute the command: make proto
  3. compile and run service: make run
  4. visit http://localhost:8080/apis/swagger/index.html in your browser, and test the CRUD api interface.

`)
			fmt.Printf("generate 'handler-pb' codes successfully, out = %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&moduleName, "module-name", "m", "", "module-name is the name of the module in the 'go.mod' file")
	//_ = cmd.MarkFlagRequired("module-name")
	cmd.Flags().StringVarP(&serverName, "server-name", "s", "", "server name")
	//_ = cmd.MarkFlagRequired("server-name")
	cmd.Flags().StringVarP(&sqlArgs.DBDsn, "db-dsn", "d", "", "db content addr, e.g. user:password@(host:port)/database")
	_ = cmd.MarkFlagRequired("db-dsn")
	cmd.Flags().StringVarP(&dbTables, "db-table", "t", "", "table name, multiple names separated by commas")
	_ = cmd.MarkFlagRequired("db-table")
	cmd.Flags().BoolVarP(&sqlArgs.IsEmbed, "embed", "e", true, "whether to embed 'gorm.Model' struct")
	cmd.Flags().IntVarP(&sqlArgs.JSONNamedType, "json-name-type", "j", 1, "json tags name type, 0:snake case, 1:camel case")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./handler-pb_<time>,"+
		" if you specify the directory where the web or microservice generated by sponge, the module-name and server-name flag can be ignored")

	return cmd
}

func runGenHandlerPbCommand(moduleName string, serverName string, codes map[string]string, outPath string) (string, error) {
	subTplName := "handler-pb"
	r := Replacers[TplNameSponge]
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	if serverName == "" {
		serverName = moduleName
	}

	// setting up template information
	subDirs := []string{"internal/model", "internal/cache", "internal/dao", "internal/ecode",
		"internal/handler", "api/serverNameExample"} // only the specified subdirectory is processed, if empty or no subdirectory is specified, it means all files
	ignoreDirs := []string{} // specify the directory in the subdirectory where processing is ignored
	ignoreFiles := []string{ // specify the files in the subdirectory to be ignored for processing
		"userExample.pb.go", "userExample.pb.validate.go", "userExample_grpc.pb.go", "userExample_router.pb.go", // api/serverNameExample
		"systemCode_http.go", "systemCode_rpc.go", "userExample_rpc.go", // internal/ecode
		"init.go", "init_test.go", // internal/model
		"handler/userExample.go", "handler/userExample_test.go", // internal/handler
		"doc.go", "cacheNameExample.go", "cacheNameExample_test.go", // internal/cache
	}

	r.SetSubDirsAndFiles(subDirs)
	r.SetIgnoreSubDirs(ignoreDirs...)
	r.SetIgnoreSubFiles(ignoreFiles...)
	fields := addHandlerPbFields(moduleName, serverName, r, codes)
	r.SetReplacementFields(fields)
	_ = r.SetOutputDir(outPath, subTplName)
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	return r.GetOutputDir(), nil
}

func addHandlerPbFields(moduleName string, serverName string, r replacer.Replacer, codes map[string]string) []replacer.Field {
	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, modelFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoTestFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, protoFile, startMark, endMark)...)
	fields = append(fields, []replacer.Field{
		{ // replace the contents of the model/userExample.go file
			Old: modelFileMark,
			New: codes[parser.CodeTypeModel],
		},
		{ // replace the contents of the dao/userExample.go file
			Old: daoFileMark,
			New: codes[parser.CodeTypeDAO],
		},
		{ // replace the contents of the v1/userExample.proto file
			Old: protoFileMark,
			New: codes[parser.CodeTypeProto],
		},
		{
			Old: selfPackageName + "/" + r.GetSourcePath(),
			New: moduleName,
		},
		{
			Old: "github.com/zhufuyi/sponge",
			New: moduleName,
		},
		// replace directory name
		{
			Old: strings.Join([]string{"api", "serverNameExample", "v1"}, gofile.GetPathDelimiter()),
			New: strings.Join([]string{"api", serverName, "v1"}, gofile.GetPathDelimiter()),
		},
		{
			Old: "api/serverNameExample/v1",
			New: fmt.Sprintf("api/%s/v1", serverName),
		},
		{
			Old: "api.serverNameExample.v1",
			New: fmt.Sprintf("api.%s.v1", strings.ReplaceAll(serverName, "-", "_")), // protobuf package no "-" signs allowed
		},
		{
			Old: "userExampleNO       = 1",
			New: fmt.Sprintf("userExampleNO = %d", rand.Intn(100)),
		},
		{
			Old: moduleName + "/pkg",
			New: "github.com/zhufuyi/sponge/pkg",
		},
		{
			Old:             "UserExamplePb",
			New:             "UserExample",
			IsCaseSensitive: true,
		},
		{
			Old: "serverNameExample",
			New: serverName,
		},
		{
			Old:             "UserExample",
			New:             codes[parser.TableName],
			IsCaseSensitive: true,
		},
	}...)

	return fields
}