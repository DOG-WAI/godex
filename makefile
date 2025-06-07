APP        := godex
TARGET     := app
ENV        := dev
INS        :=

export CGO_CXXFLAGS_ALLOW:=.*
export CGO_LDFLAGS_ALLOW:=.*
export CGO_CFLAGS_ALLOW:=.*

app:="app"

.PHONY: all test clean

# 若没有USER环境变量，从git配置中获取用户(方便windows用户)
ifeq (${USER}, )
    USER :=$(shell git config user.name)
endif

all: build
check:
ifeq ($(strip $(BK_CI_PIPELINE_NAME)),)
	@echo "\033[32m <============== 合法性校验 app ${app} =============> \033[0m"
	goimports -format-only -w -local git.code.oa.com,git.woa.com .
	gofmt -s -w .
	golangci-lint run
endif

build:
	@echo "\033[32m <============== making app ${app} =============> \033[0m"
	go build -ldflags='-w -s' $(FLAGS) -o ./${app} ./cmd

api-test: $(DEPENDENCIES)
	@echo -e "\033[32m ============== making api test =============> \033[0m"
	go test -v -coverpkg=./... ./api_test -c -o ./${app}_API.test

unit-test: $(DEPENDENCIES)
	@echo -e "\033[32m ============== making unit test =============> \033[0m"
	go test `go list ./... |grep -vE 'api_test|apitest'` -v -run='^Test' -covermode=count -gcflags=all=-l ./...

clean:
	@echo -e "\033[32m ============== cleaning files =============> \033[0m"
	rm -fv ${TARGET}

linux-app:
	@echo "\033[32m <============== making linux app ${app} =============> \033[0m"
	# 交叉编译为linux可执行文件
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s' $(FLAGS) -o ./${app} ./cmd

# 一键部署（默认上传到开发环境）
upload: linux-app
	@echo -e "\033[32m ============== uploading ${TARGET} =============> \033[0m"
	dtools bpatch -lang=go -env=${ENV} -app=${APP} -server=${TARGET} -user=${USER} -bin=${TARGET} -instances=${INS}