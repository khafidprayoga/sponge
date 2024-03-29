package generate

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zhufuyi/sponge/pkg/replacer"
	"github.com/zhufuyi/sponge/pkg/sql2code"
	"github.com/zhufuyi/sponge/pkg/sql2code/parser"
)

// HandlerCommand generate handler code
func HandlerCommand() *cobra.Command {
	var (
		moduleName string // module name for go.mod
		outPath    string // output directory
		dbTables   string // table names

		sqlArgs = sql2code.Args{
			Package:  "model",
			JSONTag:  true,
			GormType: true,
		}

		serverName     string // server name
		suitedMonoRepo bool   // whether the generated code is suitable for mono-repo
	)

	cmd := &cobra.Command{
		Use:   "handler",
		Short: "Generate handler CRUD code based on sql",
		Long: `generate handler CRUD code based on sql.

Examples:
  # generate handler code and embed gorm.model struct.
  sponge web handler --module-name=yourModuleName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user

  # generate handler code with multiple table names.
  sponge web handler --module-name=yourModuleName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=t1,t2

  # generate handler code, structure fields correspond to the column names of the table.
  sponge web handler --module-name=yourModuleName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --embed=false

  # generate handler code and specify the server directory, Note: code generation will be canceled when the latest generated file already exists.
  sponge web handler --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --out=./yourServerDir

  # if you want the generated code to suited to mono-repo, you need to specify the parameter --suited-mono-repo=true --serverName=yourServerName
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdName, srvName, smr := getNamesFromOutDir(outPath)
			if mdName != "" {
				moduleName = mdName
				serverName = srvName
				suitedMonoRepo = smr
			} else if moduleName == "" {
				return errors.New(`required flag(s) "module-name" not set, use "sponge web handler -h" for help`)
			}
			if suitedMonoRepo {
				if serverName == "" {
					return errors.New(`required flag(s) "server-name" not set, use "sponge web handler -h" for help`)
				}
				serverName = convertServerName(serverName)
				outPath = changeOutPath(outPath, serverName)
			}

			if sqlArgs.DBDriver == DBDriverMongodb {
				sqlArgs.IsEmbed = false
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

				g := &handlerGenerator{
					moduleName: moduleName,
					dbDriver:   sqlArgs.DBDriver,
					codes:      codes,
					outPath:    outPath,

					serverName:     serverName,
					suitedMonoRepo: suitedMonoRepo,
				}
				outPath, err = g.generateCode()
				if err != nil {
					return err
				}
			}

			fmt.Printf(`
using help:
  1. move the folder "internal" to your project code folder.
  2. open a terminal and execute the command: make docs
  3. compile and run service: make run
  4. visit http://localhost:8080/swagger/index.html in your browser, and test the CRUD api interface.

`)
			fmt.Printf("generate \"handler\" code successfully, out = %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&moduleName, "module-name", "m", "", "module-name is the name of the module in the go.mod file")
	//_ = cmd.MarkFlagRequired("module-name")
	cmd.Flags().StringVarP(&serverName, "server-name", "s", "", "server name")
	cmd.Flags().StringVarP(&sqlArgs.DBDriver, "db-driver", "k", "mysql", "database driver, support mysql, mongodb, postgresql, tidb, sqlite")
	cmd.Flags().StringVarP(&sqlArgs.DBDsn, "db-dsn", "d", "", "database content address, e.g. user:password@(host:port)/database. Note: if db-driver=sqlite, db-dsn must be a local sqlite db file, e.g. --db-dsn=/tmp/sponge_sqlite.db") //nolint
	_ = cmd.MarkFlagRequired("db-dsn")
	cmd.Flags().StringVarP(&dbTables, "db-table", "t", "", "table name, multiple names separated by commas")
	_ = cmd.MarkFlagRequired("db-table")
	cmd.Flags().BoolVarP(&sqlArgs.IsEmbed, "embed", "e", true, "whether to embed gorm.model struct")
	cmd.Flags().BoolVarP(&suitedMonoRepo, "suited-mono-repo", "l", false, "whether the generated code is suitable for mono-repo")
	cmd.Flags().IntVarP(&sqlArgs.JSONNamedType, "json-name-type", "j", 1, "json tags name type, 0:snake case, 1:camel case")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./handler_<time>, "+
		"if you specify the directory where the web or microservice generated by sponge, the module-name flag can be ignored")

	return cmd
}

type handlerGenerator struct {
	moduleName string
	dbDriver   string
	codes      map[string]string
	outPath    string

	serverName     string
	suitedMonoRepo bool
}

func (g *handlerGenerator) generateCode() (string, error) {
	subTplName := "handler"
	r := Replacers[TplNameSponge]
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	// setting up template information
	subDirs := []string{"internal/model", "internal/cache", "internal/dao",
		"internal/ecode", "internal/handler", "internal/routers", "internal/types"} // only the specified subdirectory is processed, if empty or no subdirectory is specified, it means all files
	ignoreDirs := []string{} // specify the directory in the subdirectory where processing is ignored
	var ignoreFiles []string
	switch strings.ToLower(g.dbDriver) {
	case DBDriverMysql, DBDriverPostgresql, DBDriverTidb, DBDriverSqlite:
		ignoreFiles = []string{ // specify the files in the subdirectory to be ignored for processing
			"systemCode_http.go", "systemCode_rpc.go", "userExample_rpc.go", // internal/ecode
			"routers.go", "routers_test.go", "routers_pbExample.go", "routers_pbExample_test.go", "userExample_router.go", // internal/routers
			"init.go", "init_test.go", "init.go.mgo", // internal/model
			"doc.go", "cacheNameExample.go", "cacheNameExample_test.go", "cache/userExample.go.mgo", // internal/cache
			"dao/userExample.go.mgo",                                                                                                              // internal/dao
			"handler/userExample_logic.go", "handler/userExample_logic_test.go", "handler/userExample.go.mgo", "handler/userExample_logic.go.mgo", // internal/handler
			"swagger_types.go", "userExample_types.go.mgo", // internal/types
		}
	case DBDriverMongodb:
		ignoreFiles = []string{ // specify the files in the subdirectory to be ignored for processing
			"systemCode_http.go", "systemCode_rpc.go", "userExample_rpc.go", // internal/ecode
			"routers.go", "routers_test.go", "routers_pbExample.go", "routers_pbExample_test.go", "userExample_router.go", // internal/routers
			"init.go", "init_test.go", "init.go.mgo", // internal/model
			"doc.go", "cacheNameExample.go", "cacheNameExample_test.go", "cache/userExample.go", "cache/userExample_test.go", // internal/cache
			"dao/userExample_test.go", "dao/userExample.go", // internal/dao
			"handler/userExample_logic.go", "handler/userExample_logic_test.go", "handler/userExample.go", "handler/userExample_test.go", "handler/userExample_logic.go.mgo", // internal/handler
			"swagger_types.go", "userExample_types.go", // internal/types
		}
	default:
		return "", errors.New("unsupported db driver: " + g.dbDriver)
	}

	r.SetSubDirsAndFiles(subDirs)
	r.SetIgnoreSubDirs(ignoreDirs...)
	r.SetIgnoreSubFiles(ignoreFiles...)
	_ = r.SetOutputDir(g.outPath, subTplName)
	fields := g.addFields(r)
	r.SetReplacementFields(fields)
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	return r.GetOutputDir(), nil
}

func (g *handlerGenerator) addFields(r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, modelFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoMgoFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoTestFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerMgoFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerTestFile, startMark, endMark)...)
	fields = append(fields, []replacer.Field{
		{ // replace the contents of the model/userExample.go file
			Old: modelFileMark,
			New: g.codes[parser.CodeTypeModel],
		},
		{ // replace the contents of the dao/userExample.go file
			Old: daoFileMark,
			New: g.codes[parser.CodeTypeDAO],
		},
		{ // replace the contents of the handler/userExample.go file
			Old: handlerFileMark,
			New: adjustmentOfIDType(g.codes[parser.CodeTypeHandler], g.dbDriver),
		},
		{
			Old: selfPackageName + "/" + r.GetSourcePath(),
			New: g.moduleName,
		},
		{
			Old: "github.com/zhufuyi/sponge",
			New: g.moduleName,
		},
		{
			Old: "userExampleNO       = 1",
			New: fmt.Sprintf("userExampleNO = %d", rand.Intn(100)),
		},
		{
			Old: g.moduleName + "/pkg",
			New: "github.com/zhufuyi/sponge/pkg",
		},
		{
			Old: "userExample_types.go.mgo",
			New: "userExample_types.go",
		},
		{
			Old: "userExample.go.mgo",
			New: "userExample.go",
		},
		{
			Old:             "UserExample",
			New:             g.codes[parser.TableName],
			IsCaseSensitive: true,
		},
	}...)

	if g.suitedMonoRepo {
		fs := SubServerCodeFields(r.GetOutputDir(), g.moduleName, g.serverName)
		fields = append(fields, fs...)
	}

	return fields
}
