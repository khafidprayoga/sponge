package generate

const (
	dockerFileHTTPCode = `# add curl, used for http service checking, can be installed without it if deployed in k8s
RUN apk add curl

COPY configs/ /app/configs/
COPY serverNameExample /app/serverNameExample
RUN chmod +x /app/serverNameExample

# http port
EXPOSE 8080`

	dockerFileGrpcCode = `# add grpc_health_probe for health check of grpc services
COPY grpc_health_probe /bin/grpc_health_probe
RUN chmod +x /bin/grpc_health_probe

COPY configs/ /app/configs/
COPY serverNameExample /app/serverNameExample
RUN chmod +x /app/serverNameExample

# grpc and http port
EXPOSE 8282 8283`

	dockerFileBuildHTTPCode = `# compressing binary files
#cd /
#upx -9 serverNameExample


# building images with binary
FROM alpine:latest
MAINTAINER zhufuyi "g.zhufuyi@gmail.com"

# set the time zone to Shanghai
RUN apk add tzdata  \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

# add curl, used for http service checking, can be installed without it if deployed in k8s
RUN apk add curl

COPY --from=build /serverNameExample /app/serverNameExample
COPY --from=build /go/src/serverNameExample/configs/serverNameExample.yml /app/configs/serverNameExample.yml

# http port
EXPOSE 8080`

	dockerFileBuildGrpcCode = `# install grpc-health-probe, for health check of grpc service
RUN go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.12
RUN cd $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-health-probe@v0.4.12 \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "all=-s -w" -o /grpc_health_probe

# compressing binary files
#cd /
#upx -9 serverNameExample
#upx -9 grpc_health_probe


# building images with binary
FROM alpine:latest
MAINTAINER zhufuyi "g.zhufuyi@gmail.com"

# set the time zone to Shanghai
RUN apk add tzdata  \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

# add grpc_health_probe for health check of grpc services
COPY --from=build /grpc_health_probe /bin/grpc_health_probe
COPY --from=build /serverNameExample /app/serverNameExample
COPY --from=build /go/src/serverNameExample/configs/serverNameExample.yml /app/configs/serverNameExample.yml

# grpc and http port
EXPOSE 8282 8283`

	imageBuildFileHTTPCode = `# compressing binary file
#cd ${DOCKERFILE_PATH}
#upx -9 ${serverName}
#cd -

echo "docker build -f ${DOCKERFILE} -t ${IMAGE_NAME_TAG} ${DOCKERFILE_PATH}"
docker build -f ${DOCKERFILE} -t ${IMAGE_NAME_TAG} ${DOCKERFILE_PATH}`

	imageBuildFileGrpcCode = `# install grpc-health-probe, for health check of grpc service
rootDockerFilePath=$(pwd)/${DOCKERFILE_PATH}
go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.12
cd $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-health-probe@v0.4.12 \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "all=-s -w" -o "${rootDockerFilePath}/grpc_health_probe"
cd -

# compressing binary file
#cd ${DOCKERFILE_PATH}
#upx -9 ${serverName}
#upx -9 grpc_health_probe
#cd -

echo "docker build -f ${DOCKERFILE} -t ${IMAGE_NAME_TAG} ${DOCKERFILE_PATH}"
docker build -f ${DOCKERFILE} -t ${IMAGE_NAME_TAG} ${DOCKERFILE_PATH}

if [ -f "${DOCKERFILE_PATH}/grpc_health_probe" ]; then
    rm -f ${DOCKERFILE_PATH}/grpc_health_probe
fi`

	imageBuildLocalFileHTTPCode = `# compressing binary file
#cd ${DOCKERFILE_PATH}
#upx -9 ${serverName}
#cd -

mkdir -p ${DOCKERFILE_PATH}/configs && cp -f configs/${serverName}.yml ${DOCKERFILE_PATH}/configs/
echo "docker build -f ${DOCKERFILE} -t ${IMAGE_NAME}:latest ${DOCKERFILE_PATH}"
docker build -f ${DOCKERFILE} -t ${IMAGE_NAME}:latest ${DOCKERFILE_PATH}`

	imageBuildLocalFileGrpcCode = `# install grpc-health-probe, for health check of grpc service
rootDockerFilePath=$(pwd)/${DOCKERFILE_PATH}
go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.12
cd $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-health-probe@v0.4.12 \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "all=-s -w" -o "${rootDockerFilePath}/grpc_health_probe"
cd -

# compressing binary file
#cd ${DOCKERFILE_PATH}
#upx -9 ${serverName}
#upx -9 grpc_health_probe
#cd -

mkdir -p ${DOCKERFILE_PATH}/configs && cp -f configs/${serverName}.yml ${DOCKERFILE_PATH}/configs/
echo "docker build -f ${DOCKERFILE} -t ${IMAGE_NAME}:latest ${DOCKERFILE_PATH}"
docker build -f ${DOCKERFILE} -t ${IMAGE_NAME}:latest ${DOCKERFILE_PATH}

if [ -f "${DOCKERFILE_PATH}/grpc_health_probe" ]; then
    rm -f ${DOCKERFILE_PATH}/grpc_health_probe
fi`

	dockerComposeFileHTTPCode = `    ports:
      - "8080:8080"   # http port
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]   # http health check, note: mirror must contain curl command`

	dockerComposeFileGrpcCode = `
    ports:
      - "8282:8282"   # grpc port
      - "8283:8283"   # grpc metrics or pprof port
    healthcheck:
      test: ["CMD", "grpc_health_probe", "-addr=localhost:8282"]    # grpc health check, note: the image must contain the grpc_health_probe command`

	k8sDeploymentFileHTTPCode = `
          ports:
            - name: http-port
              containerPort: 8080
          readinessProbe:
            httpGet:
              port: http-port
              path: /health
            initialDelaySeconds: 10
            timeoutSeconds: 2
            periodSeconds: 10
            successThreshold: 1
            failureThreshold: 3
          livenessProbe:
            httpGet:
              port: http-port
              path: /health`

	k8sDeploymentFileGrpcCode = `
          ports:
            - name: grpc-port
              containerPort: 8282
            - name: metrics-port
              containerPort: 8283
          readinessProbe:
            exec:
              command: ["/bin/grpc_health_probe", "-addr=:8282"]
            initialDelaySeconds: 10
            timeoutSeconds: 2
            periodSeconds: 10
            successThreshold: 1
            failureThreshold: 3
          livenessProbe:
            exec:
              command: ["/bin/grpc_health_probe", "-addr=:8282"]`

	k8sServiceFileHTTPCode = `  ports:
    - name: server-name-example-svc-http-port
      port: 8080
      targetPort: 8080`

	k8sServiceFileGrpcCode = `  ports:
    - name: server-name-example-svc-grpc-port
      port: 8282
      targetPort: 8282
    - name: server-name-example-svc-grpc-metrics-port
      port: 8283
      targetPort: 8283`

	configFileCode = `// code generated by https://github.com/zhufuyi/sponge

package config

import (
	"github.com/zhufuyi/sponge/pkg/conf"
)

var config *Config

func Init(configFile string, fs ...func()) error {
	config = &Config{}
	return conf.Parse(configFile, config, fs...)
}

func Show(hiddenFields ...string) string {
	return conf.Show(config, hiddenFields...)
}

func Get() *Config {
	if config == nil {
		panic("config is nil, please call config.Init() first")
	}
	return config
}

func Set(conf *Config) {
	config = conf
}
`

	configFileCcCode = `// code generated by https://github.com/zhufuyi/sponge

package config

import (
	"github.com/zhufuyi/sponge/pkg/conf"
)

func NewCenter(configFile string) (*Center, error) {
	nacosConf := &Center{}
	err := conf.Parse(configFile, nacosConf)
	return nacosConf, err
}
`

	protoShellGRPCMark = `
  # generate files *_grpc_pb.go
  protoc --proto_path=. --proto_path=./third_party \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    $allProtoFiles

  checkResult $?`

	// for rpc and rpc-pb
	protoShellServiceTmplCode = `
  moduleName=$(cat docs/gen.info | head -1 | cut -d , -f 1)
  serverName=$(cat docs/gen.info | head -1 | cut -d , -f 2)
  suitedMonoRepo=$(cat docs/gen.info | head -1 | cut -d , -f 3)

  protoc --proto_path=. --proto_path=./third_party \
    --go-rpc-tmpl_out=. --go-rpc-tmpl_opt=paths=source_relative \
    --go-rpc-tmpl_opt=moduleName=${moduleName} --go-rpc-tmpl_opt=serverName=${serverName} --go-rpc-tmpl_opt=suitedMonoRepo=${suitedMonoRepo} \
    $specifiedProtoFiles

  checkResult $?

  sponge merge rpc-pb
  checkResult $?

  colorCyan='\e[1;36m'
  highBright='\e[1m'
  markEnd='\e[0m'

  echo ""
  echo -e "${highBright}Tip:${markEnd} execute the command ${colorCyan}make run${markEnd} and then test grpc api in the file ${colorCyan}internal/service/xxx_client_test.go${markEnd}."
  echo ""`

	// for http-pb
	protoShellHandlerCode = `
  # generate the swagger document and merge all files into docs/apis.swagger.json
  protoc --proto_path=. --proto_path=./third_party \
    --openapiv2_out=. --openapiv2_opt=logtostderr=true --openapiv2_opt=allow_merge=true --openapiv2_opt=merge_file_name=docs/apis.json \
    $specifiedProtoFiles

  checkResult $?

  sponge web swagger --file=docs/apis.swagger.json
  checkResult $?

  moduleName=$(cat docs/gen.info | head -1 | cut -d , -f 1)
  serverName=$(cat docs/gen.info | head -1 | cut -d , -f 2)
  suitedMonoRepo=$(cat docs/gen.info | head -1 | cut -d , -f 3)

  protoc --proto_path=. --proto_path=./third_party \
    --go-gin_out=. --go-gin_opt=paths=source_relative --go-gin_opt=plugin=handler \
    --go-gin_opt=moduleName=${moduleName} --go-gin_opt=serverName=${serverName} --go-gin_opt=suitedMonoRepo=${suitedMonoRepo} \
    $specifiedProtoFiles

  checkResult $?

  sponge merge http-pb
  checkResult $?

  colorCyan='\e[1;36m'
  highBright='\e[1m'
  markEnd='\e[0m'

  echo ""
  echo -e "${highBright}Tip:${markEnd} execute the command ${colorCyan}make run${markEnd} and then visit ${colorCyan}http://localhost:8080/apis/swagger/index.html${markEnd} in your browser. "
  echo ""`

	// for rpc-gw
	protoShellServiceCode = `
  # Generate the swagger document and merge all files into docs/apis.swagger.json
  protoc --proto_path=. --proto_path=./third_party \
    --openapiv2_out=. --openapiv2_opt=logtostderr=true --openapiv2_opt=allow_merge=true --openapiv2_opt=merge_file_name=docs/apis.json \
    $specifiedProtoFiles

  checkResult $?

  sponge micro swagger --file=docs/apis.swagger.json
  checkResult $?

  moduleName=$(cat docs/gen.info | head -1 | cut -d , -f 1)
  serverName=$(cat docs/gen.info | head -1 | cut -d , -f 2)
  suitedMonoRepo=$(cat docs/gen.info | head -1 | cut -d , -f 3)

  protoc --proto_path=. --proto_path=./third_party \
    --go-gin_out=. --go-gin_opt=paths=source_relative --go-gin_opt=plugin=service \
    --go-gin_opt=moduleName=${moduleName} --go-gin_opt=serverName=${serverName} --go-gin_opt=suitedMonoRepo=${suitedMonoRepo} \
    $specifiedProtoFiles

  checkResult $?

  sponge merge rpc-gw-pb
  checkResult $?

  colorCyan='\e[1;36m'
  highBright='\e[1m'
  markEnd='\e[0m'

  echo ""
  echo -e "${highBright}Tip:${markEnd} execute the command ${colorCyan}make run${markEnd} and then visit ${colorCyan}http://localhost:8080/apis/swagger/index.html${markEnd} in your browser."
  echo ""`

	// for grpc-http
	protoShellServiceAndHandlerCode = `
  # generate the swagger document and merge all files into docs/apis.swagger.json
  protoc --proto_path=. --proto_path=./third_party \
    --openapiv2_out=. --openapiv2_opt=logtostderr=true --openapiv2_opt=allow_merge=true --openapiv2_opt=merge_file_name=docs/apis.json \
    $specifiedProtoFiles

  checkResult $?

  sponge web swagger --file=docs/apis.swagger.json
  checkResult $?

  moduleName=$(cat docs/gen.info | head -1 | cut -d , -f 1)
  serverName=$(cat docs/gen.info | head -1 | cut -d , -f 2)
  suitedMonoRepo=$(cat docs/gen.info | head -1 | cut -d , -f 3)

  protoc --proto_path=. --proto_path=./third_party \
    --go-rpc-tmpl_out=. --go-rpc-tmpl_opt=paths=source_relative \
    --go-rpc-tmpl_opt=moduleName=${moduleName} --go-rpc-tmpl_opt=serverName=${serverName} --go-rpc-tmpl_opt=suitedMonoRepo=${suitedMonoRepo} \
    $specifiedProtoFiles

  checkResult $?

  protoc --proto_path=. --proto_path=./third_party \
    --go-gin_out=. --go-gin_opt=paths=source_relative --go-gin_opt=plugin=handler \
    --go-gin_opt=moduleName=${moduleName} --go-gin_opt=serverName=${serverName} --go-gin_opt=suitedMonoRepo=${suitedMonoRepo} \
    $specifiedProtoFiles

  checkResult $?

  sponge merge http-pb
  checkResult $?

  colorCyan='\e[1;36m'
  highBright='\e[1m'
  markEnd='\e[0m'

  echo ""
  echo -e "${highBright}Tip:${markEnd} execute the command ${colorCyan}make run${markEnd} and then"
  echo -e "    1. test http api in your browser ${colorCyan}http://localhost:8080/apis/swagger/index.html${markEnd}"
  echo -e "    2. test grpc api in the file ${colorCyan}internal/service/xxx_client_test.go${markEnd}"
  echo ""`

	httpServerConfigCode = `# http server settings
http:
  port: 8080                # listen port
  readTimeout: 5            # read timeout, unit(second)
  writeTimeout: 5           # write timeout, unit(second), if enableHTTPProfile is true, it needs to be greater than 60s, the default value for pprof to do profiling is 60s`

	rpcServerConfigCode = `# grpc server settings
grpc:
  port: 8282                # listen port
  httpPort: 8283            # profile and metrics ports
  readTimeout: 5            # read timeout, unit(second)
  writeTimeout: 5           # write timeout, unit(second)
  enableToken: false        # whether to enable server-side token authentication, default appID=grpc, appKey=123456
  # serverSecure parameter setting
  # if type="", it means no secure connection, no need to fill in any parameters
  # if type="one-way", it means server-side certification, only the fields 'certFile' and 'keyFile' should be filled in
  # if type="two-way", it means both client and server side certification, fill in all fields
  serverSecure:
    type: ""                # secures type, "", "one-way", "two-way"
    caFile: ""              # ca certificate file, valid only in "two-way", absolute path
    certFile: ""            # server side cert file, absolute path
    keyFile: ""             # server side key file, absolute path


# grpc client-side settings, support for setting up multiple grpc clients.
grpcClient:
  - name: "your_grpc_service_name"    # grpc service name, used for service discovery
    host: "127.0.0.1"            # grpc service address, used for direct connection
    port: 8282                   # grpc service port
    registryDiscoveryType: ""    # registration and discovery types: consul, etcd, nacos, if empty, connecting to server using host and port
    enableLoadBalance: true      # whether to turn on the load balancer
    # clientSecure parameter setting
    # if type="", it means no secure connection, no need to fill in any parameters
    # if type="one-way", it means server-side certification, only the fields 'serverName' and 'certFile' should be filled in
    # if type="two-way", it means both client and server side certification, fill in all fields
    clientSecure:
      type: ""              # secures type, "", "one-way", "two-way"
      serverName: ""        # server name, e.g. *.foo.com
      caFile: ""            # client side ca file, valid only in "two-way", absolute path
      certFile: ""          # client side cert file, absolute path, if secureType="one-way", fill in server side cert file here
      keyFile: ""           # client side key file, valid only in "two-way", absolute path
    clientToken:
      enable: false         # whether to enable token authentication
      appID: ""             # app id
      appKey: ""            # app key`

	rpcGwServerConfigCode = `# http server settings
http:
  port: 8080                # listen port
  readTimeout: 5            # read timeout, unit(second)
  writeTimeout: 5           # write timeout, unit(second), if enableHTTPProfile is true, it needs to be greater than 60s, the default value for pprof to do profiling is 60s


# grpc client-side settings, support for setting up multiple grpc clients.
grpcClient:
  - name: "your_grpc_service_name"    # grpc service name, used for service discovery
    host: "127.0.0.1"            # grpc service address, used for direct connection
    port: 8282                   # grpc service port
    registryDiscoveryType: ""    # registration and discovery types: consul, etcd, nacos, if empty, connecting to server using host and port
    enableLoadBalance: true      # whether to turn on the load balancer
    # clientSecure parameter setting
    # if type="", it means no secure connection, no need to fill in any parameters
    # if type="one-way", it means server-side certification, only the fields 'serverName' and 'certFile' should be filled in
    # if type="two-way", it means both client and server side certification, fill in all fields
    clientSecure:
      type: ""              # secures type, "", "one-way", "two-way"
      serverName: ""        # server name, e.g. *.foo.com
      caFile: ""            # client side ca file, valid only in "two-way", absolute path
      certFile: ""          # client side cert file, absolute path, if secureType="one-way", fill in server side cert file here
      keyFile: ""           # client side key file, valid only in "two-way", absolute path
    clientToken:
      enable: false         # whether to enable token authentication
      appID: ""             # app id
      appKey: ""            # app key`

	grpcAndHTTPServerConfigCode = `# http server settings
http:
  port: 8080                # listen port
  readTimeout: 5            # read timeout, unit(second)
  writeTimeout: 5           # write timeout, unit(second), if enableHTTPProfile is true, it needs to be greater than 60s, the default value for pprof to do profiling is 60s


# grpc server settings
grpc:
  port: 8282                # listen port
  httpPort: 8283            # profile and metrics ports
  readTimeout: 5            # read timeout, unit(second)
  writeTimeout: 5           # write timeout, unit(second)
  enableToken: false        # whether to enable server-side token authentication, default appID=grpc, appKey=123456
  # serverSecure parameter setting
  # if type="", it means no secure connection, no need to fill in any parameters
  # if type="one-way", it means server-side certification, only the fields 'certFile' and 'keyFile' should be filled in
  # if type="two-way", it means both client and server side certification, fill in all fields
  serverSecure:
    type: ""                # secures type, "", "one-way", "two-way"
    caFile: ""              # ca certificate file, valid only in "two-way", absolute path
    certFile: ""            # server side cert file, absolute path
    keyFile: ""             # server side key file, absolute path


# grpc client-side settings, support for setting up multiple grpc clients.
grpcClient:
  - name: "your_grpc_service_name"    # grpc service name, used for service discovery
    host: "127.0.0.1"            # grpc service address, used for direct connection
    port: 8282                   # grpc service port
    registryDiscoveryType: ""    # registration and discovery types: consul, etcd, nacos, if empty, connecting to server using host and port
    enableLoadBalance: true      # whether to turn on the load balancer
    # clientSecure parameter setting
    # if type="", it means no secure connection, no need to fill in any parameters
    # if type="one-way", it means server-side certification, only the fields 'serverName' and 'certFile' should be filled in
    # if type="two-way", it means both client and server side certification, fill in all fields
    clientSecure:
      type: ""              # secures type, "", "one-way", "two-way"
      serverName: ""        # server name, e.g. *.foo.com
      caFile: ""            # client side ca file, valid only in "two-way", absolute path
      certFile: ""          # client side cert file, absolute path, if secureType="one-way", fill in server side cert file here
      keyFile: ""           # client side key file, valid only in "two-way", absolute path
    clientToken:
      enable: false         # whether to enable token authentication
      appID: ""             # app id
      appKey: ""            # app key`

	mysqlConfigCode = `# database setting
database:
  driver: "mysql"           # database driver
  # mysql settings
  mysql:
    # dsn format,  <username>:<password>@(<hostname>:<port>)/<db>?[k=v& ......]
    dsn: "root:123456@(192.168.3.37:3306)/account?parseTime=true&loc=Local&charset=utf8,utf8mb4"
    enableLog: true         # whether to turn on printing of all logs
    maxIdleConns: 10        # set the maximum number of connections in the idle connection pool
    maxOpenConns: 100       # set the maximum number of open database connections
    connMaxLifetime: 30     # sets the maximum time for which the connection can be reused, in minutes
    #slavesDsn:             # sets slaves mysql dsn, array type
    #  - "your slave dsn 1"
    #  - "your slave dsn 2"
    #mastersDsn:            # sets masters mysql dsn, array type, non-required field, if there is only one master, there is no need to set the mastersDsn field, the default dsn field is mysql master.
    #  - "your master dsn`

	postgresqlConfigCode = `database:
  driver: "postgresql"      # database driver
  # postgresql settings
  postgresql:
    # dsn format,  <username>:<password>@<hostname>:<port>/<db>?[k=v& ......]
    dsn: "root:123456@192.168.3.37:5432/account?sslmode=disable"
    enableLog: true         # whether to turn on printing of all logs
    maxIdleConns: 10        # set the maximum number of connections in the idle connection pool
    maxOpenConns: 100       # set the maximum number of open database connections
    connMaxLifetime: 30     # sets the maximum time for which the connection can be reused, in minutes`

	sqliteConfigCode = `database:
  driver: "sqlite"      # database driver
  # sqlite settings
  sqlite:
    dbFile: "test/sql/sqlite/sponge.db"
    enableLog: true         # whether to turn on printing of all logs
    maxIdleConns: 10        # set the maximum number of connections in the idle connection pool
    maxOpenConns: 100       # set the maximum number of open database connections
    connMaxLifetime: 30     # sets the maximum time for which the connection can be reused, in minutes`

	mongodbConfigCode = `database:
  driver: "mongodb"      # database driver
  # mongodb settings
  mongodb:
    # dsn format,  <username>:<password>@<hostname1>:<port1>[,<hostname2>:<port2>,......]/<db>?[k=v& ......]
    # parameter k=v see https://www.mongodb.com/docs/drivers/go/current/fundamentals/connections/connection-guide/#connection-options
    dsn: "root:123456@192.168.3.37:27017/account?connectTimeoutMS=15000"`

	undeterminedDatabaseConfigCode = `# set database configuration. reference-db-config-url
database:
  driver: "mysql"           # database driver
  # mysql settings
  mysql:
    # dsn format,  <username>:<password>@(<hostname>:<port>)/<db>?[k=v& ......]
    dsn: "root:123456@(192.168.3.37:3306)/account?parseTime=true&loc=Local&charset=utf8,utf8mb4"
    enableLog: true         # whether to turn on printing of all logs
    maxIdleConns: 10        # set the maximum number of connections in the idle connection pool
    maxOpenConns: 100       # set the maximum number of open database connections
    connMaxLifetime: 30     # sets the maximum time for which the connection can be reused, in minutes
`

	modelInitDBFileMysqlCode = `// InitDB connect database
func InitDB() {
	switch strings.ToLower(config.Get().Database.Driver) {
	case ggorm.DBDriverMysql, ggorm.DBDriverTidb:
		InitMysql()
	default:
		panic("InitDB error, unsupported database driver: " + config.Get().Database.Driver)
	}
}

// InitMysql connect mysql
func InitMysql() {
	opts := []ggorm.Option{
		ggorm.WithMaxIdleConns(config.Get().Database.Mysql.MaxIdleConns),
		ggorm.WithMaxOpenConns(config.Get().Database.Mysql.MaxOpenConns),
		ggorm.WithConnMaxLifetime(time.Duration(config.Get().Database.Mysql.ConnMaxLifetime) * time.Minute),
	}
	if config.Get().Database.Mysql.EnableLog {
		opts = append(opts,
			ggorm.WithLogging(logger.Get()),
			ggorm.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, ggorm.WithEnableTrace())
	}

	// setting mysql slave and master dsn addresses,
	// if there is no read/write separation, you can comment out the following piece of code
	opts = append(opts, ggorm.WithRWSeparation(
		config.Get().Database.Mysql.SlavesDsn,
		config.Get().Database.Mysql.MastersDsn...,
	))

	// add custom gorm plugin
	//opts = append(opts, ggorm.WithGormPlugin(yourPlugin))

	var dsn = utils.AdaptiveMysqlDsn(config.Get().Database.Mysql.Dsn)
	var err error
	db, err = ggorm.InitMysql(dsn, opts...)
	if err != nil {
		panic("InitMysql error: " + err.Error())
	}
}`

	modelInitDBFilePostgresqlCode = `// InitDB connect database
func InitDB() {
	switch strings.ToLower(config.Get().Database.Driver) {
	case ggorm.DBDriverPostgresql:
		InitPostgresql()
	default:
		panic("InitDB error, unsupported database driver: " + config.Get().Database.Driver)
	}
}

// InitPostgresql connect postgresql
func InitPostgresql() {
	opts := []ggorm.Option{
		ggorm.WithMaxIdleConns(config.Get().Database.Postgresql.MaxIdleConns),
		ggorm.WithMaxOpenConns(config.Get().Database.Postgresql.MaxOpenConns),
		ggorm.WithConnMaxLifetime(time.Duration(config.Get().Database.Postgresql.ConnMaxLifetime) * time.Minute),
	}
	if config.Get().Database.Postgresql.EnableLog {
		opts = append(opts,
			ggorm.WithLogging(logger.Get()),
			ggorm.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, ggorm.WithEnableTrace())
	}

	// add custom gorm plugin
	//opts = append(opts, ggorm.WithGormPlugin(yourPlugin))

	var dsn = utils.AdaptivePostgresqlDsn(config.Get().Database.Postgresql.Dsn)
	var err error
	db, err = ggorm.InitPostgresql(dsn, opts...)
	if err != nil {
		panic("InitPostgresql error: " + err.Error())
	}
}`

	modelInitDBFileSqliteCode = `// InitDB connect database
func InitDB() {
	switch strings.ToLower(config.Get().Database.Driver) {
	case ggorm.DBDriverSqlite:
		InitSqlite()
	default:
		panic("InitDB error, unsupported database driver: " + config.Get().Database.Driver)
	}
}

// InitSqlite connect sqlite
func InitSqlite() {
	opts := []ggorm.Option{
		ggorm.WithMaxIdleConns(config.Get().Database.Sqlite.MaxIdleConns),
		ggorm.WithMaxOpenConns(config.Get().Database.Sqlite.MaxOpenConns),
		ggorm.WithConnMaxLifetime(time.Duration(config.Get().Database.Sqlite.ConnMaxLifetime) * time.Minute),
	}
	if config.Get().Database.Sqlite.EnableLog {
		opts = append(opts,
			ggorm.WithLogging(logger.Get()),
			ggorm.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, ggorm.WithEnableTrace())
	}

	var err error
	var dbFile = utils.AdaptiveSqlite(config.Get().Database.Sqlite.DBFile)
	db, err = ggorm.InitSqlite(dbFile, opts...)
	if err != nil {
		panic("InitSqlite error: " + err.Error())
	}
}`

	embedTimeCode = `value.CreatedAt = record.CreatedAt.Format(time.RFC3339)
	value.UpdatedAt = record.UpdatedAt.Format(time.RFC3339)`
)
