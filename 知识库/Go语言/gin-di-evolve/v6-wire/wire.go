//go:build wireinject
// +build wireinject

package main

// ============================================================
// wire.go：Injector 声明（只在生成时编译）
// ============================================================
//
// 本文件带 wireinject 构建标签：
//   - 平时 go run / go build 不会编进这个文件
//   - 运行 wire 命令时才会用来生成 wire_gen.go
//
// 重新生成（需已安装 wire）：
//
//	go install github.com/google/wire/cmd/wire@latest
//	cd v6-wire
//	wire
//
// 本目录已提交 wire_gen.go，没有 wire 命令也能直接 go run 。

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// InitializeApp Wire 入口：声明「我要一棵 *gin.Engine」，并列出所有 Provider。
// 函数体里的 return nil 只是占位，生成后会被换成真正的组装代码。
func InitializeApp() *gin.Engine {
	wire.Build(
		newMemDB,
		NewUserRepository,
		NewSessionRepository,
		NewProjectRepository,
		NewAuthService,
		NewProjectService,
		provideDependencies,
		NewRouter,
	)
	return nil
}
